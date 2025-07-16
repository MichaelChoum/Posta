package logic

import (
	"context"
	"encoding/json"
	"posta/application/like/mq/internal/svc"
	"posta/application/like/mq/internal/types"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

const maxRetry = 3 // 防止消息丢失的最大重试次数
type LikeActionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeActionLogic {
	return &LikeActionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 应该是Comsume是kafka interface的一个method
func (l *LikeActionLogic) Consume(ctx context.Context, key, val string) error {
	// 反序列化
	var msg types.LikeActionMsg
	err := json.Unmarshal([]byte(val), &msg)

	if err != nil {
		l.Logger.Errorf("[LikeComsumer] Json.Unmarshal err:%v", err)
		return nil
	}

	// 为防止消费失败，这里多次重试
	for i := 0; i < maxRetry; i++ {
		err := l.hanleLikeAction(l.ctx, l.svcCtx, &msg)
		if err != nil {
			l.Logger.Errorf("[LikeComsumer] hanleLikeAction err:%v, times %d", err, i+1)
			time.Sleep(100 * time.Millisecond)
		} else {
			return nil
		}
	}

	l.Logger.Errorf("[LikeComsumer] permanently failed: %v", msg)
	return nil
}

func (l *LikeActionLogic) hanleLikeAction(ctx context.Context, svcCtx *svc.ServiceContext, msg *types.LikeActionMsg) error {
	if msg.LikeAction == 0 {
		_, err := svcCtx.LikeModel.InsertIgnore(ctx, msg.BizId, msg.ObjId, msg.UserId)
		if err != nil {
			l.Logger.Errorf("LikeActionMsg InsertIgnore err:%v", err)
			return err
		}
	} else {
		err := svcCtx.LikeModel.DeleteByBizObjUser(ctx, msg.BizId, msg.ObjId, msg.UserId)
		if err != nil {
			l.Logger.Errorf("LikeActionMsg DeleteByBizObjUser err:%v", err)
			return err
		}
	}
	return nil
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.KqConsumerConf, NewLikeActionLogic(ctx, svcCtx)), // NewLikeActionLogic 要实现 ConsumeBatch
	}
}
