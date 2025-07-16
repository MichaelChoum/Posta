package logic

import (
	"context"
	"errors"
	"posta/application/like/rpc/internal/model"
	"posta/application/like/rpc/internal/svc"
	"posta/application/like/rpc/pb"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCountLogic {
	return &LikeCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询点赞数（单个）
func (l *LikeCountLogic) LikeCount(in *pb.LikeCountRequest) (*pb.LikeCountResponse, error) {

	likeCountKey := LikeCountKey(in.BizId, in.ObjId)

	// 查询缓存，存储的count是string
	count, err := l.svcCtx.BizRedis.GetCtx(l.ctx, likeCountKey)
	if err != nil {
		l.Logger.Errorf("redis %v GetCtx error: %v", likeCountKey, err)
		return nil, err
	}
	// value == "" && err == nil表示key不存在，但操作是正常的
	if count != "" {
		countInt, err := strconv.ParseInt(count, 10, 64)
		if err != nil {
			l.Logger.Errorf("redis %v strconv.ParseInt error: %v", likeCountKey, err)
		}
		return &pb.LikeCountResponse{Count: countInt}, nil
	}

	// 再查询数据库
	record, err := l.svcCtx.LikeCountModel.FindOneByBizIdObjId(l.ctx, in.BizId, in.ObjId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 没有数据视为 0
			// 防止缓存穿透
			_ = l.svcCtx.BizRedis.SetexCtx(l.ctx, likeCountKey, strconv.FormatInt(int64(0), 10), 7*LikesExpire)
			return &pb.LikeCountResponse{Count: 0}, nil
		}
		l.Logger.Errorf("mysql like_count query error: %v", err)
		return nil, err
	}

	// 写入缓存
	_ = l.svcCtx.BizRedis.SetexCtx(l.ctx, likeCountKey, strconv.FormatInt(int64(record.LikeNum), 10), 7*LikesExpire)

	return &pb.LikeCountResponse{Count: record.LikeNum}, nil
}
