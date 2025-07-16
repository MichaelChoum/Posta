package logic

import (
	"cmp"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/mr"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/threading"
	"math"
	"posta/application/reply/rpc/internal/code"
	"posta/application/reply/rpc/internal/model"
	"posta/application/reply/rpc/internal/types"
	"slices"
	"strconv"
	"time"

	"posta/application/reply/rpc/internal/svc"
	"posta/application/reply/rpc/service"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	prefixFirstReplies  = "biz#firstReplies#%d#%d"
	prefixSecondReplies = "biz#secondReplies#%d#%d"
	// 两天的时间
	repliesExpire = 3600 * 24 * 2
)

type RepliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RepliesLogic {
	return &RepliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 可以查看文章的一级评论，也可以查看一级评论下的二级评论
func (l *RepliesLogic) Replies(in *service.RepliesRequest) (*service.RepliesResponse, error) {
	if in.SortType != types.SortPublishTime && in.SortType != types.SortLikeCount {
		return nil, code.SortTypeInvalid
	}
	if in.TargetId <= 0 {
		return nil, code.ArticleIdInvalid
	}
	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}
	if in.Cursor == 0 {
		if in.SortType == types.SortPublishTime {
			if in.ParentId == 0 {
				// 一级评论按时间排序是从新到旧，相当于开启一个新的话题
				in.Cursor = time.Now().Unix()
			} else {
				// 二级评论按时间排序是从旧到新，相当于看讨论过程
				in.Cursor = 1
			}
		} else {
			in.Cursor = types.DefaultSortLikeCursor
		}
	}

	var (
		sortField       string
		sortLikeNum     int64
		sortPublishTime string
	)

	if in.SortType == types.SortLikeCount {
		sortField = "like_num"
		sortLikeNum = in.Cursor
	} else {
		sortField = "create_time"
		// 将时间从int64的cursor转化为string形式
		sortPublishTime = time.Unix(in.Cursor, 0).Format("2006-01-02 15:04:05")
	}

	var (
		err            error
		isCache, isEnd bool
		lastId, cursor int64
		// []*T 表示 “T 类型的指针组成的切片”，切片可以理解为动态数组
		curPage []*service.ReplyItem
		// 这是数据库的一条行记录的指针
		replies []*model.Reply
	)

	replyIds, _ := l.cacheReplies(l.ctx, in.TargetId, in.ParentId, in.Cursor, in.PageSize, in.SortType)

	// 这里的in.ReplyId != replyIds[0]是为了防止缓存中本来有数据
	if len(replyIds) > 0 && in.ReplyId != replyIds[0] {
		isCache = true
		if replyIds[len(replyIds)-1] == -1 || replyIds[len(replyIds)-1] == math.MaxInt64 {
			if in.SortType == types.SortLikeCount || in.ParentId == 0 {
				isEnd = true
			} else {
				// 如果是查看一级评论下的二级评论，那么必须保证redis的zset中-1对应的score为replyIds[len(replyIds)-2])对应的score才意味着读取完了。
				// 注意：理论上应该再对比一下id的，毕竟两条评论的createtime可能相同。
				publishTimeKey := secondRepliesKey(in.ParentId, types.SortPublishTime)

				// mysqlMax表示mysql存储最大score
				mysqlMax, _ := l.svcCtx.BizRedis.ZscoreCtx(l.ctx, publishTimeKey, strconv.Itoa(int(replyIds[len(replyIds)-1])))
				// redisMax表示redis的zset中存储的member不为MaxInt64的最大score
				redisMax, _ := l.svcCtx.BizRedis.ZscoreCtx(l.ctx, publishTimeKey, strconv.Itoa(int(replyIds[len(replyIds)-2])))
				if len(replyIds) > 2 && redisMax == mysqlMax {
					isEnd = true
				}
			}
		}

		// goroutine加速查询，但返回的replies是无序的，
		replies, err = l.replyByIds(l.ctx, replyIds)
		if err != nil {
			return nil, err
		}

		// 再进行排序
		var cmpFunc func(a, b *model.Reply) int
		if sortField == "like_num" {
			cmpFunc = func(a, b *model.Reply) int {
				// 大的在前，表示降序（点赞从高到低）
				return cmp.Compare(b.LikeNum, a.LikeNum)
			}
		} else if in.ParentId == 0 {
			cmpFunc = func(a, b *model.Reply) int {
				return cmp.Compare(b.CreateTime.Unix(), a.CreateTime.Unix())
			}
		} else {
			cmpFunc = func(a, b *model.Reply) int {
				return cmp.Compare(a.CreateTime.Unix(), b.CreateTime.Unix())
			}
		}
		slices.SortFunc(replies, cmpFunc)

		for _, reply := range replies {
			curPage = append(curPage, &service.ReplyItem{
				Id:            reply.Id,
				ReplyUserId:   reply.ReplyUserId,
				BeReplyUserId: reply.BeReplyUserId,
				ParentId:      reply.ParentId,
				Content:       reply.Content,
				LikeCount:     reply.LikeNum,
				CreateTime:    reply.CreateTime.Unix(),
			})
		}
	} else {
		var (
			v   interface{}
			err error
		)
		// 如果是根据文章id查一级评论。
		if in.ParentId == 0 {
			// SingleFlight（请求合并）机制，避免同一个 key 被并发重复请求数据库，只执行一次，其他请求等待复用结果。
			// 但是SingleFlight控制范围为单个进程内，多个goroutine并发请求。
			// key：fmt.Sprintf("ArticlesBuUserId:%d:%d", in.TargetId, in.SortType)
			// 函数体：要被合并执行的函数逻辑
			v, err, _ = l.svcCtx.SingleFlightGroup.Do(fmt.Sprintf("FirstRepliesByArticleId:%d:%d", in.TargetId, in.SortType), func() (interface{}, error) {
				// 这里DefaultLimit设置得比较大为200，可以提前加载进缓存
				return l.svcCtx.ReplyModel.FirstRepliesByArticleId(l.ctx, in.TargetId, sortLikeNum, sortPublishTime, sortField, types.DefaultLimit)
			})
		} else {
			v, err, _ = l.svcCtx.SingleFlightGroup.Do(fmt.Sprintf("SecondRepliesByFristReplyId:%d:%d", in.ParentId, in.SortType), func() (interface{}, error) {
				return l.svcCtx.ReplyModel.SecondRepliesByFirstReplyId(l.ctx, in.ParentId, sortLikeNum, sortPublishTime, sortField, types.DefaultLimit)
			})
		}

		if err != nil {
			logx.Errorf("RepliesByArticleId error: %d sortField: %s error: %v", in.TargetId, sortField, err)
			return nil, err
		}
		if v == nil {
			return &service.RepliesResponse{}, nil
		}
		replies = v.([]*model.Reply)
		var firstPageReplies []*model.Reply
		if len(replies) > int(in.PageSize) {
			firstPageReplies = replies[:int(in.PageSize)]
		} else {
			firstPageReplies = replies
			isEnd = true
		}

		// 这里虽然给BeReplyUserId赋值了，但前端保证BeReplyUserId为0时不显示就可以了。
		for _, reply := range firstPageReplies {
			curPage = append(curPage, &service.ReplyItem{
				Id:            reply.Id,
				ReplyUserId:   reply.ReplyUserId,
				BeReplyUserId: reply.BeReplyUserId,
				ParentId:      reply.ParentId,
				Content:       reply.Content,
				LikeCount:     reply.LikeNum,
				CreateTime:    reply.CreateTime.Unix(),
			})
		}
	}

	// 更新cursor和lastId
	if len(curPage) > 0 {
		pageLast := curPage[len(curPage)-1]
		lastId = pageLast.Id
		if in.SortType == types.SortPublishTime {
			cursor = pageLast.CreateTime
		} else {
			cursor = pageLast.LikeCount
		}
		if cursor < 0 {
			cursor = 0
		}

		for k, reply := range curPage {
			if in.SortType == types.SortPublishTime {
				// 遍历 curPage 找到与上一页最后一条 相同 Cursor + 相同 Id 的那条评论；
				// 从它的下标 k 开始往后截取 curPage，丢弃它之前的内容（包括它本身）；
				// 用 curPage[k:] 作为新的结果，避免重复展示这条评论。
				if reply.CreateTime == in.Cursor && reply.Id == in.ReplyId {
					curPage = curPage[k:]
					break
				}
			} else {
				if reply.LikeCount == in.Cursor && reply.Id == in.ReplyId {
				}
				curPage = curPage[k:]
				break
			}
		}
	}

	ret := &service.RepliesResponse{
		IsEnd:   isEnd,
		Cursor:  cursor,
		ReplyId: lastId,
		Replies: curPage,
	}

	if !isCache {
		// 异步写缓存
		threading.GoSafe(func() {
			ctx := context.Background() // Debug: 使用新上下文，避免主流程已退出
			var err error

			// 对于一级评论按时间排序
			if len(replies) < types.DefaultLimit && len(replies) > 0 {
				// 注意：这里加入redis中的是zset的东西，只是id而已。省去了排序过程，但是没有将整个查到的数据都放进redis中，避免redis数据量太大了。
				// 之前考虑到可能存在“假尾”问题，但是这里的逻辑是len(replies) < types.DefaultLimit时才会在redis中缓存-1。
				// 如果大于的话，那就仅仅是将这些数据加入redis，并不设置结束符号，所以没问题。
				if in.SortType == types.SortLikeCount || in.ParentId == 0 {
					replies = append(replies, &model.Reply{Id: -1})
				}
			}

			// 对于二级评论按时间排序
			if in.SortType == types.SortPublishTime && in.ParentId != 0 {
				maxTime, err := l.svcCtx.ReplyModel.MaxCreateTimeByParentId(ctx, in.ParentId)
				if err != nil {
					logx.Errorf("MaxCreateTimeByParentId for parentId %d error: %v", in.ParentId, err)
				} else {
					replies = append(replies, &model.Reply{Id: math.MaxInt64, CreateTime: maxTime})
				}
			}

			err = l.addCacheReplies(ctx, replies, in.TargetId, in.ParentId, in.SortType)
			if err != nil {
				logx.Errorf("addCacheReplies error: %v", err)
			}
		})
	}

	return ret, nil
}

// 从 Redis 缓存中，获取某个文章的一级评论ID列表（按某种排序方式）或者一级评论的二级ID列表，用于分页展示。
func (l *RepliesLogic) cacheReplies(ctx context.Context, articleid, parentid, cursor, ps int64, sortType int32) ([]int64, error) {
	var key string
	if parentid == 0 {
		key = firstRepliesKey(articleid, sortType)
	} else {
		key = secondRepliesKey(articleid, sortType)
	}
	b, err := l.svcCtx.BizRedis.ExistsCtx(ctx, key)
	if err != nil {
		logx.Errorf("ExistsCtx key: %s error: %v", key, err)
	}

	// 注意：更新缓存缓存过期时间，防止缓存击穿。
	if b {
		err = l.svcCtx.BizRedis.ExpireCtx(ctx, key, repliesExpire)
		if err != nil {
			logx.Errorf("ExpireCtx key: %s error: %v", key, err)
		}
	}

	// 按分数从大到小（ZREVRANGEBYSCORE）取分数在 [0, cursor] 范围内的元素
	// 返回元素和它的分数
	// 只返回分页内的元素（偏移量 0，数量 ps）
	var pairs []redis.Pair
	if parentid != 0 && sortType == types.SortPublishTime {
		// 一级评论下的二级评论，按时间排序，应该是时间越旧分数越低。所以这里应该是从低往高读。
		pairs, err = l.svcCtx.BizRedis.ZrangebyscoreWithScoresAndLimitCtx(ctx, key, cursor, math.MaxInt64, 0, int(ps))
	} else {
		pairs, err = l.svcCtx.BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, 0, cursor, 0, int(ps))
	}
	if err != nil {
		logx.Errorf("ZrevrangebyscoreWithScoresAndLimit key: %s error: %v", key, err)
		return nil, err
	}

	var ids []int64
	for _, pair := range pairs {
		// redis的key只能为字符串，需要转化为int64
		id, err := strconv.ParseInt(pair.Key, 10, 64)
		if err != nil {
			logx.Errorf("strconv.ParseInt key: %s error: %v", pair.Key, err)
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// 接收一个 articleIds 的整数 ID 列表，然后并发查询每一个文章详情，最后组合成列表返回。
// 用法：
// 参数含义
// 假设调用形式：
// mr.MapReduce[In, MapOut, ReduceOut](
//
//	func(source chan<- In) { ... },         // 输入生产者，把输入放到 channel
//	func(in In, writer mr.Writer[MapOut], cancel func(error)) { ... },  // map 阶段，处理每个输入，写结果
//	func(pipe <-chan MapOut, writer mr.Writer[ReduceOut], cancel func(error)) { ... }, // reduce 阶段，合并所有 map 结果
//
// )
// 第1个函数：负责产生输入数据（往 source channel 里写所有输入）
// 第2个函数（map）：从 source 读输入，多个 goroutine 并发执行，对每个输入并行处理，产生结果。
// 第3个函数（reduce）：从 map 的结果流中读数据，进行汇总操作，输出最终结果
func (l *RepliesLogic) replyByIds(ctx context.Context, replyIds []int64) ([]*model.Reply, error) {
	// [a, b, c]输入类型，行记录类型，输出
	// chan int64：一个可以发送和接收int64 的通道。
	// chan<- int64：一个只能发送int64 的通道（写通道）。
	// <-chan int64：一个只能接收int64 的通道（读通道）。
	replies, err := mr.MapReduce[int64, *model.Reply, []*model.Reply](func(source chan<- int64) {
		// 把每个 replyId 发到 source channel，作为每个 Map 的输入。
		// 第一次出现，用:=。声明并赋值。
		for _, rid := range replyIds {
			if rid == -1 || rid == math.MaxInt64 {
				continue
			}
			source <- rid
		}
	}, func(id int64, writer mr.Writer[*model.Reply], cancel func(error)) {
		// 对每个 replyId：
		// 调用 FindOne 从数据库或缓存查一条评论
		// 如果成功，把结果写入结果流
		// 如果失败，调用 cancel(err) 终止整个流程
		p, err := l.svcCtx.ReplyModel.FindOne(ctx, id)
		if err != nil {
			cancel(err)
			return
		}
		writer.Write(p)
	}, func(pipe <-chan *model.Reply, writer mr.Writer[[]*model.Reply], cancel func(error)) {
		// 从 Map 阶段输出的所有评论中读取
		// 把它们合并成一个 []*model.Reply 切片
		// 返回给调用方
		var replies []*model.Reply
		for reply := range pipe {
			replies = append(replies, reply)
		}
		writer.Write(replies)
	})

	if err != nil {
		return nil, err
	}
	return replies, nil
}

func (l *RepliesLogic) addCacheReplies(ctx context.Context, replies []*model.Reply, articleId, parentId int64, sortType int32) error {
	if len(replies) == 0 {
		return nil
	}
	var key string
	if parentId == 0 {
		key = firstRepliesKey(articleId, sortType)
	} else {
		key = secondRepliesKey(parentId, sortType)
	}

	for _, reply := range replies {
		var score int64
		if sortType == types.SortLikeCount {
			score = reply.LikeNum
		} else if sortType == types.SortPublishTime {
			score = reply.CreateTime.Local().Unix()
		}
		// 对于id为-1的虚拟评论，score可能为负值，比如time.time的0值转成Unix。
		if score < 0 {
			score = 0
		}
		// itoa 将 int 转成字符串（string）
		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, key, score, strconv.Itoa(int(reply.Id)))

		if err != nil {
			return err
		}
	}
	return l.svcCtx.BizRedis.ExpireCtx(ctx, key, repliesExpire)
}

func firstRepliesKey(articleid int64, sortType int32) string {
	return fmt.Sprintf(prefixFirstReplies, articleid, sortType)
}

func secondRepliesKey(replyid int64, sortType int32) string {
	return fmt.Sprintf(prefixSecondReplies, replyid, sortType)
}
