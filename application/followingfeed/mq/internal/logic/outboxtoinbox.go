package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/threading"
	"posta/application/followingfeed/mq/internal/model"
	"posta/application/followingfeed/mq/internal/svc"
	"posta/application/followingfeed/mq/internal/types"
	"strconv"
	"time"
)

const (
	prefixInbox = "biz#inbox#%d"
	inboxExpire = 3600 * 24 * 2
)

type OutboxToInboxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOutboxToInboxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OutboxToInboxLogic {
	return &OutboxToInboxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OutboxToInboxLogic) Consume(ctx context.Context, _, val string) error {
	logx.Infof("Consume msg val: %s", val)
	var msg *types.CanalArticleMsg
	err := json.Unmarshal([]byte(val), &msg)
	if err != nil {
		logx.Errorf("Consume val: %s error: %v", val, err)
		return err
	}

	return l.articleOperate(msg)
}

func (l *OutboxToInboxLogic) articleOperate(msg *types.CanalArticleMsg) error {
	if len(msg.Data) == 0 {
		return nil
	}

	for _, d := range msg.Data {
		status, _ := strconv.Atoi(d.Status)
		articleId, _ := strconv.ParseInt(d.ID, 10, 64)
		authorId, _ := strconv.ParseInt(d.AuthorId, 10, 64)
		t, err := time.ParseInLocation("2006-01-02 15:04:05", d.PublishTime, time.Local)

		// 这里沿用了b站的称呼，称为小up主。
		var isSmallUp bool
		fansCount, err := l.svcCtx.FollowCountModel.GetFansCount(l.ctx, authorId)

		if err != nil {
			l.Logger.Errorf("GetFansCount fail, author_id = %d, err: %v", authorId, err)
			continue
		}

		// 如果粉丝数小于BigUpFansThreshold，则是小up主
		if fansCount < types.BigUpFansThreshold {
			isSmallUp = true
		}

		switch status {
		// 如果是发布了一篇新文章
		case types.ArticleStatusVisible:
			if !isSmallUp {
				continue
			}
			fanIds, err := l.svcCtx.FollowModel.GetFanIds(l.ctx, authorId)
			if err != nil {
				l.Logger.Errorf("GetFanIds fail, author_id = %d, err: %v", authorId, err)
				continue
			}
			if len(fanIds) == 0 {
				continue
			}

			// 开启独立协程，避免阻塞主线程
			threading.GoSafe(func() {
				inboxBatch := make([]*model.UserInbox, 0, types.BatchSize)
				insertedInboxes := make([]*model.UserInbox, 0, len(fanIds))
				// 这里使用小批量插入，更加安全。
				for i, fanId := range fanIds {
					inboxBatch = append(inboxBatch, &model.UserInbox{
						UserId:      fanId,
						SenderId:    authorId,
						ArticleId:   articleId,
						Status:      types.ArticleStatusVisible,
						PublishTime: t,
						IsRead:      0,
					})
					if len(inboxBatch) == types.BatchSize || i == len(fanIds)-1 {
						_, err := l.svcCtx.UserInBoxModel.BatchInsert(context.Background(), inboxBatch)
						if err != nil {
							l.Logger.Errorf("BatchInsert fail, author_id = %d, err: %v", authorId, err)
						} else {
							insertedInboxes = append(insertedInboxes, inboxBatch...)
						}
						inboxBatch = inboxBatch[:0] // 重置
					}
				}

				// 更新redis
				l.updateInboxCacheForInsert(insertedInboxes)
			})

		// 如果是删除了一篇文章
		case types.ArticleStatusUserDelete:
			if !isSmallUp {
				continue
			}

			threading.GoSafe(func() {
				// 数据库批量更新为删除
				// 如果是物理删除，建议写一个 BatchDelete 方法
				// 如果是逻辑删除(状态字段)，可以批量 update
				_, err := l.svcCtx.UserInBoxModel.BatchSetDeleteByArticle(context.Background(), articleId)
				if err != nil {
					l.Logger.Errorf("BatchSetDeleteByArticle fail, article_id = %d, err: %v", articleId, err)
				}

				// 更新redis
				l.updateInboxCacheForDelete(authorId, articleId)
			})
		}
	}
	return nil
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.ArticleKqConsumerConf, NewOutboxToInboxLogic(ctx, svcCtx)),
		kq.MustNewQueue(svcCtx.Config.FollowKqConsumerConf, NewFollowCompensateLogic(ctx, svcCtx)),
	}
}

// 更新缓存：插入新文章到粉丝收信箱
func (l *OutboxToInboxLogic) updateInboxCacheForInsert(inboxes []*model.UserInbox) {
	for _, inbox := range inboxes {
		key := l.inboxKey(inbox.UserId)

		// 检查缓存是否存在
		exists, err := l.svcCtx.BizRedis.ExistsCtx(context.Background(), key)
		if err != nil {
			l.Logger.Errorf("Check cache exists error: %v", err)
			continue
		}

		if exists {
			// 缓存存在，添加新的收信箱记录
			score := inbox.PublishTime.Local().Unix()
			inboxId, err := l.svcCtx.UserInBoxModel.FindOneByUserIdSenderIdArticleId(context.Background(), inbox.UserId, inbox.SenderId, inbox.ArticleId)
			_, err = l.svcCtx.BizRedis.ZaddCtx(context.Background(), key, score, strconv.Itoa(int(inboxId)))
			if err != nil {
				l.Logger.Errorf("Update inbox cache for insert error: %v", err)
			}

			// 刷新过期时间
			l.svcCtx.BizRedis.ExpireCtx(context.Background(), key, inboxExpire)
		}
		// 如果缓存不存在，不主动创建，等下次查询时懒记载
	}
}

func (l *OutboxToInboxLogic) inboxKey(uid int64) string {
	return fmt.Sprintf(prefixInbox, uid)
}

// 更新缓存：删除文章从所有相关收信箱
func (l *OutboxToInboxLogic) updateInboxCacheForDelete(authorId, articleId int64) {
	// 获取作者的所有粉丝
	fanIds, err := l.svcCtx.FollowModel.GetFanIds(context.Background(), authorId)
	if err != nil {
		l.Logger.Errorf("GetFanIds for cache delete error: %v", err)
		return
	}

	// 为每个粉丝的收信箱缓存中移除该文章的记录
	for _, fanId := range fanIds {
		key := l.inboxKey(fanId)

		// 检查缓存是否存在
		exists, err := l.svcCtx.BizRedis.ExistsCtx(context.Background(), key)
		if err != nil {
			l.Logger.Errorf("Check cache exists error: %v", err)
			continue
		}

		if exists {
			// 需要找到对应的inbox记录并删除
			// 由于我们在缓存中存储的是inbox.Id而不是articleId，需要查找对应关系
			l.removeArticleFromInboxCache(fanId, articleId)
		}
	}
}

// 从特定用户的收信箱缓存中移除特定文章
func (l *OutboxToInboxLogic) removeArticleFromInboxCache(userId, articleId int64) {
	key := l.inboxKey(userId)

	// 获取该用户所有的收信箱记录ID
	pairs, err := l.svcCtx.BizRedis.ZrevrangeWithScoresCtx(context.Background(), key, 0, -1)
	if err != nil {
		l.Logger.Errorf("Get inbox cache for removal error: %v", err)
		return
	}

	// 查找并删除对应文章的inbox记录
	for _, pair := range pairs {
		inboxId, err := strconv.ParseInt(pair.Key, 10, 64)
		if err != nil {
			continue
		}

		// 查询inbox记录确认是否为要删除的文章
		inbox, err := l.svcCtx.UserInBoxModel.FindOne(context.Background(), inboxId)
		if err != nil {
			continue
		}

		if inbox.ArticleId == articleId {
			// 从缓存中删除这条记录
			_, err = l.svcCtx.BizRedis.ZremCtx(context.Background(), key, pair.Key)
			if err != nil {
				l.Logger.Errorf("Remove from inbox cache error: %v", err)
			}
			break
		}
	}
}
