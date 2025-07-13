package logic

import (
	"context"
	"math"
	"posta/application/reply/rpc/internal/code"
	"posta/application/reply/rpc/internal/model"
	"posta/application/reply/rpc/internal/types"
	"strconv"
	"time"

	"posta/application/reply/rpc/internal/svc"
	"posta/application/reply/rpc/service"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReplyPublishLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReplyPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplyPublishLogic {
	return &ReplyPublishLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReplyPublishLogic) ReplyPublish(in *service.ReplyPublishRequest) (*service.ReplyPublishResponse, error) {
	if in.ReplyUserId <= 0 {
		return nil, code.UserIdInvalid
	}
	if len(in.Content) == 0 {
		return nil, code.ReplyContentEmpty
	}
	reply := &model.Reply{
		ReplyUserId:   in.ReplyUserId,
		TargetId:      in.TargetId,
		BeReplyUserId: in.BeReplyUserId,
		ParentId:      in.ParentId,
		RootReplyId:   in.RootReplyId,
		Content:       in.Content,
		Status:        types.ReplyStatusOk,
		CreateTime:    time.Now(),
	}
	ret, err := l.svcCtx.ReplyModel.Insert(l.ctx, reply)

	if err != nil {
		l.Logger.Errorf("Reply Insert req: %v error: %v", in, err)
		return nil, err
	}

	replyId, err := ret.LastInsertId()
	if err != nil {
		l.Logger.Errorf("LastInsertId error: %v", in, err)
	}

	// 注意：为了保证缓存和数据库的一致性，只有当缓存存在时，才会往缓存中加入数据。
	// 如果缓存不存在，说明没人调用过articles方法，我们只需要改变数据库就行，如果仍然执行zadd的话那缓存中就只有这个值了。
	// 这里分了两类讨论，分别是in.ParentId为0，即一级评论；或是in.ParentId不为0，即二级评论。
	if in.ParentId == 0 {
		var (
			publishTimeScore = reply.CreateTime.Unix()
			replyIdStr       = strconv.FormatInt(replyId, 10)
			publishTimeKey   = firstRepliesKey(in.TargetId, types.SortPublishTime)
			likeNumKey       = firstRepliesKey(in.TargetId, types.SortLikeCount)
		)
		// 如果 Redis ZSet 存在
		b, _ := l.svcCtx.BizRedis.ExistsCtx(l.ctx, publishTimeKey)
		if b {
			_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, publishTimeKey, publishTimeScore, replyIdStr)
			if err != nil {
				l.Logger.Errorf("ZaddCtx req: %v error: %v", in, err)
			}
		}
		b, _ = l.svcCtx.BizRedis.ExistsCtx(l.ctx, likeNumKey)
		if b {
			_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, likeNumKey, 0, replyIdStr)
			if err != nil {
				logx.Errorf("ZaddCtx req: %v error: %v", in, err)
			}
		}
	} else {
		var (
			replyIdStr     = strconv.FormatInt(replyId, 10)
			publishTimeKey = secondRepliesKey(in.ParentId, types.SortPublishTime)
			likeNumKey     = secondRepliesKey(in.ParentId, types.SortLikeCount)
		)
		// 注意：这里只有一级评论才需要将最新评论加入到zset中，因为是按时间排是从新到旧。而二级评论是从旧到新，如果将最新评论加入到zset中会有问题。
		b, _ := l.svcCtx.BizRedis.ExistsCtx(l.ctx, publishTimeKey)
		if b {
			// 查询当前缓存中 "math.MaxInt64" 的 score（可能没有）。
			existScore, err := l.svcCtx.BizRedis.ZscoreCtx(l.ctx, publishTimeKey, strconv.Itoa(math.MaxInt64))
			// 如果redis中没有缓存"math.MaxInt64"，则从数据库里去查。
			if err != nil {
				existTime, errDB := l.svcCtx.ReplyModel.MaxCreateTimeByParentId(l.ctx, in.ParentId)
				if errDB != nil {
					l.Logger.Errorf("Reply findmaxcreatetimebyparentid error: %v", errDB)
				}
				existScore = existTime.Unix()
			}

			newScore := reply.CreateTime.Unix()

			// 只有当新评论时间更大时，才更新math.MaxInt64标记
			if newScore > existScore {
				// 在zSet中，Member是唯一的如果 "math.MaxInt64" 已经存在于 ZSet 中，它会被自动更新为新的 newScore，无需你手动删除原来的。
				_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, publishTimeKey, newScore, strconv.Itoa(math.MaxInt64))
				if err != nil {
					l.Logger.Errorf("ZaddCtx update MaxInt64 for key %s failed: %v", publishTimeKey, err)
				}
			} else {
				_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, publishTimeKey, existScore, strconv.Itoa(math.MaxInt64))
				if err != nil {
					l.Logger.Errorf("ZaddCtx update MaxInt64 for key %s failed: %v", publishTimeKey, err)
				}
			}
		}
		b, _ = l.svcCtx.BizRedis.ExistsCtx(l.ctx, likeNumKey)
		if b {
			_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, likeNumKey, 0, replyIdStr)
			if err != nil {
				logx.Errorf("ZaddCtx req: %v error: %v", in, err)
			}
		}
	}

	return &service.ReplyPublishResponse{ReplyId: replyId}, nil
}
