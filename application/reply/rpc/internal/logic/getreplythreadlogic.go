package logic

import (
	"context"
	"posta/application/reply/rpc/internal/code"
	"posta/application/reply/rpc/internal/types"
	"time"

	"posta/application/reply/rpc/internal/svc"
	"posta/application/reply/rpc/service"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReplyThreadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReplyThreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReplyThreadLogic {
	return &GetReplyThreadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 可以查看二级评论下的对话链。
func (l *GetReplyThreadLogic) GetReplyThread(in *service.GetReplyThreadRequest) (*service.GetReplyThreadResponse, error) {
	// 基本逻辑是找到同一个root_reply_id的所有评论，并按照时间顺序从旧到排序。
	// 因为这个功能相比于查看一级评论和查看二级评论用的比较少，这里就不再存储redis缓存。事实上，知乎和抖音也不存在这个功能，只有b站实现了，一定程度上可以反映大家对这个需求不高。
	// 但我个人比较喜欢这个功能，所以这里将它实现了。需要时直接查询数据库。

	if in.RootReplyId <= 0 {
		return nil, code.ReplyIdInvalid
	}

	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}

	if in.Cursor == 0 {
		in.Cursor = 1
	}

	sortPublishTime := time.Unix(in.Cursor, 0).Format("2006-01-02 15:04:05")
	replies, err := l.svcCtx.ReplyModel.RepliesByRootReplyId(l.ctx, in.RootReplyId, sortPublishTime, types.DefaultLimit)
	if err != nil {
		logx.Errorf("RepliesByRootReplyId %d error: %v", in.GetRootReplyId, err)
		return nil, err
	}
	if replies == nil {
		return &service.GetReplyThreadResponse{}, nil
	}

	var (
		isEnd          bool
		cursor, lastId int64
		curPage        []*service.ReplyItem
	)
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

	if len(replies) < int(in.PageSize) {
		isEnd = true
	}

	if len(curPage) > 0 {
		pageLast := curPage[len(curPage)-1]
		cursor = pageLast.CreateTime
		lastId = pageLast.Id
		if cursor < 0 {
			cursor = 0
		}
	}

	ret := &service.GetReplyThreadResponse{
		IsEnd:   isEnd,
		Cursor:  cursor,
		ReplyId: lastId,
		Replies: curPage,
	}

	return ret, nil
}
