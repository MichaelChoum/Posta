package logic

import (
	"context"
	"encoding/json"
	"posta/application/user/rpc/user"

	"posta/application/applet/internal/svc"
	"posta/application/applet/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo() (*types.UserInfoResponse, error) {
	userId, err := l.ctx.Value(types.UserIdKey).(json.Number).Int64()
	if err != nil {
		return nil, err
	}
	if userId == 0 {
		return nil, nil
	}
	u, err := l.svcCtx.UserRPC.FindById(l.ctx, &user.FindByIdRequest{
		UserId: uint64(userId),
	})
	if err != nil {
		logx.Errorf("FindById userId: %d error: %v", userId, err)
		return nil, err
	}

	return &types.UserInfoResponse{
		UserId:   int64(u.UserId),
		Username: u.Username,
		Avatar:   u.Avatar,
	}, nil
}
