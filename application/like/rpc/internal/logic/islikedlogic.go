package logic

import (
	"context"
	"errors"
	"posta/application/like/rpc/internal/model"
	"posta/application/like/rpc/internal/svc"
	"posta/application/like/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsLikedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsLikedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsLikedLogic {
	return &IsLikedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户是否点赞（单个）
func (l *IsLikedLogic) IsLiked(in *pb.IsLikedRequest) (*pb.IsLikedResponse, error) {

	likeRecordKey := LikeRecordKey(in.BizId, in.UserId)

	// 查询缓存中是否存在
	exist, err := l.svcCtx.BizRedis.SismemberCtx(l.ctx, likeRecordKey, in.ObjId)
	if err != nil {
		l.Logger.Error("sismember %s Sismember error: %v", likeRecordKey, err)
		return nil, err
	}

	// 如果缓存中能找到
	if exist {
		return &pb.IsLikedResponse{Liked: true}, nil
	}

	// 否则，再从数据库中找
	_, err = l.svcCtx.LikeModel.FindOneByBizIdObjIdUserId(l.ctx, in.BizId, in.ObjId, in.UserId)
	if err != nil {
		// 没有找到也是error
		if errors.Is(err, model.ErrNotFound) {
			return &pb.IsLikedResponse{Liked: false}, nil
		}
		l.Logger.Error("sismember %s LikeModel error: %v", likeRecordKey, err)
		return nil, err
	}

	// 从数据库中读到点赞记录，写入缓存中
	_, err = l.svcCtx.BizRedis.SaddCtx(l.ctx, likeRecordKey, in.UserId)
	if err != nil {
		l.Logger.Errorf("redis %v Sadd error: %v", likeRecordKey, err)
	}
	err = l.svcCtx.BizRedis.ExpireCtx(l.ctx, likeRecordKey, LikesExpire)
	if err != nil {
		l.Logger.Errorf("redis %v Expire error: %v", likeRecordKey, err)
	}

	return &pb.IsLikedResponse{Liked: true}, nil
}
