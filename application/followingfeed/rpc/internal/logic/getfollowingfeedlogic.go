package logic

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/mr"
	"github.com/zeromicro/go-zero/core/threading"
	"posta/application/followingfeed/rpc/internal/model"
	"posta/application/followingfeed/rpc/internal/types"
	"sort"
	"strconv"
	"time"

	"posta/application/followingfeed/rpc/internal/svc"
	"posta/application/followingfeed/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	prefixInbox = "biz#inbox#%d"
	inboxExpire = 3600 * 24 * 2
)

type GetFollowingFeedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFollowingFeedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowingFeedLogic {
	return &GetFollowingFeedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFollowingFeedLogic) GetFollowingFeed(in *pb.GetFollowingRequest) (*pb.GetFollowingResponse, error) {
	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}

	if in.CursorSmallUp == 0 {
		in.CursorSmallUp = time.Now().Unix()
	}
	if in.CursorBigUp == 0 {
		in.CursorBigUp = time.Now().Unix()
	}

	// 1. 获取用户的关注列表，分离大UP
	bigUpIds, err := l.getBigUpIds(l.ctx, in.UserId)
	if err != nil {
		l.Logger.Errorf("getBigUpIds - error: %v", err)
		return nil, err
	}

	// 2. 从缓存或数据库获取PageSize条小UP收信箱id（这部分类似于articleslogic）
	smallUpFeedLites, _, cursorSmallUp, lastIdSmallUp, err := l.fetchInbox(l.ctx, in.UserId, in.CursorSmallUp, in.InboxId, in.PageSize)
	if err != nil {
		l.Logger.Errorf("fetchInbox - error: %v", err)
	}

	// 3. 获取大UP动态（直接查缓存或数据库），
	bigUpFeedLites, _, cursorBigUp, lastIdBigUp, err := l.fetchOutboxForBigUps(bigUpIds, in.CursorBigUp, in.ArticleId, in.PageSize)
	if err != nil {
		logx.Errorf("failed to fetch big UP outbox: %v", err)
	}

	// 4. 合并排序大小UP的动态，并获取排序后的前pagesize条数据
	finalFeed, isEnd := l.mergeAndFetchDetails(
		smallUpFeedLites, bigUpFeedLites, in.PageSize)

	// 封装返回结果
	resp := &pb.GetFollowingResponse{
		FollowingItems: finalFeed,
		IsEnd:          isEnd,
		CursorSmallUp:  cursorSmallUp,
		InboxId:        lastIdSmallUp,
		CursorBigUp:    cursorBigUp,
		ArticleId:      lastIdBigUp,
	}
	return resp, nil
}

func (l *GetFollowingFeedLogic) getBigUpIds(ctx context.Context, userId int64) (bigUpIds []int64, err error) {
	followedIds, err := l.svcCtx.FollowModel.GetFollowedIds(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("getBigUpIds - error: %v", err)
		return nil, err
	}

	for _, followedId := range followedIds {
		fansCount, _ := l.svcCtx.FollowCountModel.GetFansCount(l.ctx, followedId)
		if fansCount >= types.BigUpFansThreshold {
			bigUpIds = append(bigUpIds, followedId)
		}
	}
	return bigUpIds, nil
}

// 输出是articleid，而不是完整的article信息。
// 缓存的是userinboxid，而不是articleid。
func (l *GetFollowingFeedLogic) fetchInbox(ctx context.Context, userId, incursor, inboxId, pageSize int64) ([]types.InboxItemLite, bool, int64, int64, error) {

	var (
		isCache, isEnd    bool
		lastId, outcursor int64
		userInboxs        []*model.UserInbox
		curPage           []types.InboxItemLite
	)

	publishTime := time.Unix(incursor, 0).Format("2006-01-02 15:04:05")

	smallUpFeedIds, err := l.cacheInbox(l.ctx, userId, incursor, pageSize)
	if len(smallUpFeedIds) > 0 {
		isCache = true
		if smallUpFeedIds[len(smallUpFeedIds)-1] == -1 {
			isEnd = true
		}

		// 返回的userinboxs是无序的，不过没关系，后面还要跟大up的一起排序。
		userInboxs, err = l.userinboxByIds(l.ctx, smallUpFeedIds)
		if err != nil {
			return nil, false, 0, 0, err
		}

		// 收信箱中文章的id
		for _, userInbox := range userInboxs {
			curPage = append(curPage, types.InboxItemLite{
				Id:          userInbox.Id,
				ArticleId:   userInbox.ArticleId,
				PublishTime: userInbox.PublishTime.Unix(),
			})
		}

	} else {
		userInboxs, err = l.svcCtx.UserInBoxModel.UserInboxsByUserId(l.ctx, userId, publishTime, types.DefaultLimit)

		if err != nil {
			logx.Errorf("UserInboxsByUserId userId: %d error: %v", userId, err)
			return nil, false, 0, 0, err
		}

		var firstPageUserInboxs []*model.UserInbox
		if len(userInboxs) > int(pageSize) {
			firstPageUserInboxs = userInboxs[:pageSize]
		} else {
			firstPageUserInboxs = userInboxs
			isEnd = true
		}
		for _, userInbox := range firstPageUserInboxs {
			curPage = append(curPage, types.InboxItemLite{
				Id:          userInbox.Id,
				ArticleId:   userInbox.ArticleId,
				PublishTime: userInbox.PublishTime.Unix(),
			})
		}

	}

	if len(curPage) > 0 {
		pageLast := curPage[len(curPage)-1]
		lastId = pageLast.Id
		outcursor = pageLast.PublishTime
		if outcursor < 0 {
			outcursor = 0
		}
		for k, inboxItemLite := range curPage {
			if inboxItemLite.PublishTime == incursor && inboxItemLite.Id == inboxId {
				curPage = curPage[k:]
				break
			}
		}
	}

	if !isCache {
		threading.GoSafe(func() {
			if len(userInboxs) < types.DefaultLimit && len(userInboxs) > 0 {
				userInboxs = append(userInboxs, &model.UserInbox{Id: -1})
			}
			err = l.addCacheUserInbox(context.Background(), userInboxs, userId)
			if err != nil {
				logx.Errorf("addCacheUserInbox error: %v", err)
			}
		})
	}

	return curPage, isEnd, outcursor, lastId, nil
}

func (l *GetFollowingFeedLogic) cacheInbox(ctx context.Context, userId, cursor, pageSize int64) ([]int64, error) {
	key := inboxKey(userId)
	b, err := l.svcCtx.BizRedis.ExistsCtx(ctx, key)
	if err != nil {
		l.Logger.Errorf("exist cacheInbox key: %s error: %v", key, err)
	}
	// 注意：更新缓存缓存过期时间，防止缓存击穿。
	if b {
		err = l.svcCtx.BizRedis.ExpireCtx(ctx, key, inboxExpire)
		if err != nil {
			l.Logger.Errorf("expire cacheInbox key: %s error: %v", key, err)
		}
	}
	pairs, err := l.svcCtx.BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, 0, cursor, 0, int(pageSize))
	if err != nil {
		l.Logger.Errorf("zrevrangebyscoreWithScoresAndLimitCtx key: %s, error: %v", key, err)
		return nil, err
	}
	var ids []int64
	for _, pair := range pairs {
		id, err := strconv.ParseInt(pair.Key, 10, 64)
		if err != nil {
			l.Logger.Errorf("strconv.ParseInt key: %s error: %v", pair.Key, err)
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func inboxKey(uid int64) string {
	return fmt.Sprintf(prefixInbox, uid)
}

// 接收一个 inboxIds 的整数 ID 列表，然后并发查询每一个inbox详情，最后组合成列表返回。
func (l *GetFollowingFeedLogic) userinboxByIds(ctx context.Context, userinboxIds []int64) ([]*model.UserInbox, error) {
	userinboxs, err := mr.MapReduce[int64, *model.UserInbox, []*model.UserInbox](func(source chan<- int64) {
		for _, userinboxid := range userinboxIds {
			if userinboxid == -1 {
				continue
			}
			source <- userinboxid
		}
	}, func(id int64, writer mr.Writer[*model.UserInbox], cancel func(error)) {
		p, err := l.svcCtx.UserInBoxModel.FindOne(ctx, id)
		if err != nil {
			cancel(err)
			return
		}
		writer.Write(p)
	}, func(pipe <-chan *model.UserInbox, writer mr.Writer[[]*model.UserInbox], cancel func(error)) {
		var userinboxs []*model.UserInbox
		for userinbox := range pipe {
			userinboxs = append(userinboxs, userinbox)
		}
		writer.Write(userinboxs)
	})
	if err != nil {
		return nil, err
	}

	return userinboxs, nil
}

// 将一批文章的信息（ID + 排序依据）加入到 Redis 的有序集合（ZSet）中用于缓存，并设置过期时间。
func (l *GetFollowingFeedLogic) addCacheUserInbox(ctx context.Context, userinboxs []*model.UserInbox, userId int64) error {
	if len(userinboxs) == 0 {
		return nil
	}
	key := inboxKey(userId)
	for _, userinbox := range userinboxs {
		var score int64
		if userinbox.Id != -1 {
			score = userinbox.PublishTime.Local().Unix()
		}
		if score < 0 {
			score = 0
		}
		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, key, score, strconv.Itoa(int(userinbox.Id)))
		if err != nil {
			return err
		}
	}

	return l.svcCtx.BizRedis.ExpireCtx(ctx, key, inboxExpire)
}

func (l *GetFollowingFeedLogic) fetchOutboxForBigUps(userIds []int64, inCursor, articleId, pageSize int64) ([]types.ArticleLite, bool, int64, int64, error) {
	// 初始化变量
	var (
		articleLites  []types.ArticleLite // 用于存储最终的大UP文章列表
		isEnd         bool                // 是否数据已经到底
		nextCursor    int64               // 下一页游标对应的时间戳
		nextArticleId int64               // 下一页游标对应的文章ID（tie-breaking）
	)

	articleLites, err := l.svcCtx.ArticleModel.ArticlesLiteByUserIds(l.ctx, userIds, inCursor, pageSize)

	if err != nil {
		return nil, false, 0, 0, err
	}

	// 如果结果数量小于 pageSize，说明到底了
	if len(articleLites) < int(pageSize) {
		isEnd = true
	}

	// cursor相同时，利用ArticleId去重
	for k, articleLite := range articleLites {
		if articleLite.PublishTime == inCursor && articleLite.ArticleId == articleId {
			articleLites = articleLites[k:]
			break
		}
	}

	// 下一页游标：取最后一条的发布信息（包括时间和文章ID）
	lastArticle := articleLites[len(articleLites)-1]
	nextCursor = lastArticle.PublishTime
	nextArticleId = lastArticle.ArticleId

	return articleLites, isEnd, nextCursor, nextArticleId, nil
}

func (l *GetFollowingFeedLogic) mergeAndFetchDetails(smallUpFeed []types.InboxItemLite, bigUpFeed []types.ArticleLite, pageSize int64) ([]*pb.GetFollowingItem, bool) {
	type FeedItem struct {
		ArticleId   int64
		PublishTime int64
		IsSmallUp   bool  // 区分数据来源，true 表示来自小UP，false 表示来自大UP
		Id          int64 // 针对小UP，需要 InboxId 时用，针对大UP为 0
	}
	// 1. 合并大小UP的动态，并按时间排序
	allFeed := make([]FeedItem, 0, len(smallUpFeed)+len(bigUpFeed))

	for _, item := range smallUpFeed {
		allFeed = append(allFeed, FeedItem{
			ArticleId:   item.ArticleId,
			PublishTime: item.PublishTime,
			IsSmallUp:   true,
			Id:          item.Id,
		})
	}

	for _, item := range bigUpFeed {
		allFeed = append(allFeed, FeedItem{
			ArticleId:   item.ArticleId,
			PublishTime: item.PublishTime,
			IsSmallUp:   false,
			Id:          0,
		})
	}

	sort.Slice(allFeed, func(i, j int) bool {
		// 首先按发布时间降序排序
		if allFeed[i].PublishTime == allFeed[j].PublishTime {
			// 如果时间相同，按文章ID降序排序（防止游标冲突）
			return allFeed[i].ArticleId > allFeed[j].ArticleId
		}
		return allFeed[i].PublishTime > allFeed[j].PublishTime
	})

	// 2. 截取前 pageSize 条动态
	selectedFeed := allFeed
	isEnd := false
	if len(allFeed) < int(pageSize) {
		isEnd = true
	} else {
		selectedFeed = allFeed[:pageSize]
	}

	// 3. 根据文章ID获取文章详细信息
	var articles []*model.Article
	for _, item := range selectedFeed {
		article, err := l.svcCtx.ArticleModel.FindOne(l.ctx, item.ArticleId)
		if err != nil {
			l.Logger.Error("get article error: %v", err)
			continue
		}
		articles = append(articles, article)
	}

	// 4. 将最终的数据封装成 GetFollowingItem
	finalFeed := make([]*pb.GetFollowingItem, 0, len(articles))
	for _, article := range articles {
		finalFeed = append(finalFeed, &pb.GetFollowingItem{
			Id:           article.Id,
			Title:        article.Title,
			Content:      article.Content,
			Description:  article.Description,
			Cover:        article.Cover,
			CommentCount: article.CommentNum,
			LikeCount:    article.LikeNum,
			PublishTime:  article.PublishTime.Local().Unix(),
			AuthorId:     article.AuthorId,
		})
	}

	return finalFeed, isEnd
}
