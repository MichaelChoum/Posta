package logic

import (
	"context"
	"fmt"

	"posta/application/like/mq/internal/svc"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type ThumbupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewThumbupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThumbupLogic {
	return &ThumbupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注意：这里多加了一个ctx context.Context，不然会出错，也不知道为什么
func (l *ThumbupLogic) Consume(ctx context.Context, key, val string) error {
	fmt.Printf("get key: %s val: %s\n", key, val)
	return nil
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.KqConsumerConf, NewThumbupLogic(ctx, svcCtx)),
	}
}
