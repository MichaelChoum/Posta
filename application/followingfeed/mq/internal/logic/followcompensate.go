package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"posta/application/followingfeed/mq/internal/model"
	"posta/application/followingfeed/mq/internal/svc"
	"posta/application/followingfeed/mq/internal/types"
	"strconv"
)

// 这个文件主要负责关注关系的收信箱补充，比如新增了一个关注，取消了关注。

type FollowCompensateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowCompensateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowCompensateLogic {
	return &FollowCompensateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FollowCompensateLogic) Consume(ctx context.Context, _, val string) error {
	logx.Infof("Consume msg val: %s", val)
	var msg *types.CanalFollowMsg
	err := json.Unmarshal([]byte(val), &msg)
	if err != nil {
		logx.Errorf("Consume val: %s error: %v", val, err)
		return err
	}

	return l.followOperate(msg)
}

func (l *FollowCompensateLogic) followOperate(msg *types.CanalFollowMsg) error {
	if len(msg.Data) == 0 {
		return nil
	}

	for _, d := range msg.Data {
		fanId, _ := strconv.ParseInt(d.UserId, 10, 64)
		followedUserId, _ := strconv.ParseInt(d.FollowedUserID, 10, 64)
		status, _ := strconv.Atoi(d.Status)

		// 这里沿用了b站的称呼，称为小up主。
		var isSmallUp bool
		fansCount, err := l.svcCtx.FollowCountModel.GetFansCount(l.ctx, followedUserId)

		if err != nil {
			l.Logger.Errorf("GetFansCount fail, author_id = %d, err: %v", followedUserId, err)
			continue
		}

		// 如果粉丝数小于BigUpFansThreshold，则是小up主
		if fansCount < types.BigUpFansThreshold {
			isSmallUp = true
		}

		// 新增关注
		if status == types.FollowStatusFollow {
			// 如果不是小up，不用补偿进收信箱
			if !isSmallUp {
				continue
			}

			// 查找小up的所有articleId
			articles, err := l.svcCtx.ArticleModel.ArticlesByUserId(l.ctx, followedUserId)
			if err != nil {
				l.Logger.Errorf("ArticleIdsByUserId fail, author_id = %d, err: %v", followedUserId, err)
			}

			if len(articles) == 0 {
				continue
			}

			// 准备批量插入
			inboxBatch := make([]*model.UserInbox, 0, len(articles))
			for _, article := range articles {
				inboxBatch = append(inboxBatch, &model.UserInbox{
					UserId:      fanId,
					SenderId:    followedUserId,
					ArticleId:   article.Id,
					Status:      types.ArticleStatusVisible,
					PublishTime: article.PublishTime,
					IsRead:      0,
				})
			}

			_, err = l.svcCtx.UserInBoxModel.BatchInsert(l.ctx, inboxBatch)
			if err != nil {
				l.Logger.Errorf("BatchInsert fail, author_id = %d, err: %v", followedUserId, err)
			} else {
				// 更新缓存，这里直接把缓存全删了，因为关注是低频操作，这样代价也不大，代码逻辑更清晰
				l.updateInboxCacheForFollowCompensate(fanId)
			}
		}
		if status == types.FollowStatusUnfollow {
			if !isSmallUp {
				continue
			}
			_, err := l.svcCtx.UserInBoxModel.BatchSetDeleteBySenRec(l.ctx, fanId, followedUserId)
			if err != nil {
				l.Logger.Errorf("BatchSetDeleteBySenRec error: %v", err)
			} else {
				l.updateInboxCacheForUnfollow(fanId, followedUserId)
			}

		}
	}
	return nil
}

func (l *FollowCompensateLogic) inboxKey(uid int64) string {
	return fmt.Sprintf(prefixInbox, uid)
}

// 关注补偿：更新收信箱缓存
// 这里采用了缓存实效策略，如果新增关注，则直接删除缓存。
// 重新加载，避免了复杂的同步逻辑。
// 性能可接受：关注/取消关注是低频操作。
func (l *FollowCompensateLogic) updateInboxCacheForFollowCompensate(fanId int64) {
	key := l.inboxKey(fanId)

	// 直接删除缓存，让下次查询重新构建
	exists, err := l.svcCtx.BizRedis.ExistsCtx(context.Background(), key)
	if err != nil {
		l.Logger.Errorf("Check cache exists error: %v", err)
		return
	}

	if exists {
		_, err = l.svcCtx.BizRedis.DelCtx(context.Background(), key)
		if err != nil {
			l.Logger.Errorf("Invalidate inbox cache error: %v", err)
		} else {
			l.Logger.Infof("Invalidated inbox cache for user %d, will rebuild on next query", fanId)
		}
	}
}

// 取消关注：从收信箱缓存中删除特定UP的所有文章
func (l *FollowCompensateLogic) updateInboxCacheForUnfollow(fanId, unfollowedUserId int64) {
	key := l.inboxKey(fanId)

	// 检查缓存是否存在
	exists, err := l.svcCtx.BizRedis.ExistsCtx(context.Background(), key)
	if err != nil {
		l.Logger.Errorf("Check cache exists error: %v", err)
		return
	}

	if !exists {
		return // 缓存不存在，无需处理
	}

	// 获取该用户收信箱缓存中所有记录
	pairs, err := l.svcCtx.BizRedis.ZrevrangeWithScoresCtx(context.Background(), key, 0, -1)
	if err != nil {
		l.Logger.Errorf("Get inbox cache for unfollow error: %v", err)
		return
	}

	// 查找并删除来自被取消关注用户的所有文章
	toRemove := make([]any, 0)
	for _, pair := range pairs {
		inboxId, err := strconv.ParseInt(pair.Key, 10, 64)
		if err != nil {
			continue
		}

		// 查询inbox记录确认发送者
		inbox, err := l.svcCtx.UserInBoxModel.FindOne(context.Background(), inboxId)
		if err != nil {
			continue
		}

		if inbox.SenderId == unfollowedUserId {
			toRemove = append(toRemove, pair.Key)
		}
	}

	// 批量删除缓存中的记录
	if len(toRemove) > 0 {
		_, err = l.svcCtx.BizRedis.ZremCtx(context.Background(), key, toRemove...)
		if err != nil {
			l.Logger.Errorf("Remove from inbox cache for unfollow error: %v", err)
		} else {
			l.Logger.Infof("Removed %d inbox items from cache for user %d unfollowing %d",
				len(toRemove), fanId, unfollowedUserId)
		}
	}
}
