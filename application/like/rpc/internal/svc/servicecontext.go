package svc

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"posta/application/like/rpc/internal/config"
	"posta/application/like/rpc/internal/model"
)

type ServiceContext struct {
	Config         config.Config
	LikeModel      model.LikeRecordModel
	LikeCountModel model.LikeCountModel
	BizRedis       *redis.Redis
	KqPusherClient *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds, _ := redis.NewRedis(redis.RedisConf{
		Host:     c.BizRedis.Host,
		Pass:     c.BizRedis.Pass,
		Type:     c.BizRedis.Type,
		NonBlock: true,
	})

	return &ServiceContext{
		Config:         c,
		LikeModel:      model.NewLikeRecordModel(sqlx.NewMysql(c.DataSource)),
		LikeCountModel: model.NewLikeCountModel(sqlx.NewMysql(c.DataSource)),
		BizRedis:       rds,
		KqPusherClient: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
	}
}
