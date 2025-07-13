package logic

import (
	"context"
	"math"
	"posta/application/reply/rpc/internal/code"
	"posta/application/reply/rpc/internal/types"
	"posta/pkg/xcode"
	"strconv"

	"posta/application/reply/rpc/internal/svc"
	"posta/application/reply/rpc/service"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReplyDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReplyDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplyDeleteLogic {
	return &ReplyDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReplyDeleteLogic) ReplyDelete(in *service.ReplyDeleteRequest) (*service.ReplyDeleteResponse, error) {
	if in.ReplyUserId <= 0 {
		return nil, code.UserIdInvalid
	}
	if in.ReplyId <= 0 {
		return nil, code.ReplyIdInvalid
	}
	reply, err := l.svcCtx.ReplyModel.FindOne(l.ctx, in.ReplyId)

	if err != nil {
		l.Logger.Errorf("ReplyDelet FindOne req: %v error: %v", in, err)
	}

	if reply.ReplyUserId != in.ReplyUserId {
		return nil, xcode.AccessDenied
	}

	err = l.svcCtx.ReplyModel.UpdateReplyStatus(l.ctx, in.ReplyId, types.ReplyStatusDelete)
	if err != nil {
		l.Logger.Errorf("UpdateReplyStatus req: %v error: %v", in, err)
		return nil, err
	}
	// 注意：删除文章要保证数据库和缓存的一致性。在上面操作完数据库之后，这里删除缓存中的数据。
	// 这里是删除对应文章的一级评论
	if in.ParentId == 0 {
		_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, firstRepliesKey(in.TargetId, types.SortPublishTime), strconv.Itoa(int(in.ReplyId)))
		if err != nil {
			l.Logger.Errorf("ZremRedis req: %v error: %v", in, err)
		}
		_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, firstRepliesKey(in.TargetId, types.SortLikeCount), strconv.Itoa(int(in.ReplyId)))
		if err != nil {
			l.Logger.Errorf("ZremRedis req: %v error: %v", in, err)
		}
	} else { // 这里是删除对应一级评论的二级评论
		_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, secondRepliesKey(in.ParentId, types.SortPublishTime), strconv.Itoa(int(in.ReplyId)))
		if err != nil {
			l.Logger.Errorf("ZremRedis req: %v error: %v", in, err)
		}

		// 判断是否需要删除 "maxInt64" 这个标记
		// 1. 查询 Redis 中 "maxInt64" 的 score
		latestScore, err := l.svcCtx.BizRedis.ZscoreCtx(l.ctx, secondRepliesKey(in.ParentId, types.SortPublishTime), strconv.Itoa(math.MaxInt64))
		if err != nil {
			l.Logger.Errorf("ZscoreCtx 'maxint64' error: %v", err)
		}

		// 2. 查询该评论的 create_time
		reply, err := l.svcCtx.ReplyModel.FindOne(l.ctx, in.ReplyId)
		if err != nil {
			l.Logger.Errorf("FindOne error: %v", err)
		}

		// 3. 如果相同，则删除
		if latestScore == reply.CreateTime.Unix() {
			_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, secondRepliesKey(in.ParentId, types.SortPublishTime), strconv.Itoa(math.MaxInt64))
			if err != nil {
				l.Logger.Errorf("ZremRedis remove 'maxint64' error: %v", err)
			}
		}

		_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, secondRepliesKey(in.ParentId, types.SortLikeCount), in.ReplyId)
		if err != nil {
			l.Logger.Errorf("ZremRedis req: %v error: %v", in, err)
		}
	}

	return &service.ReplyDeleteResponse{}, nil
}
