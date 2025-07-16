package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"posta/application/like/mq/internal/config"
	"posta/application/like/mq/internal/model"
)

type ServiceContext struct {
	Config         config.Config
	LikeModel      model.LikeRecordModel
	LikeCountModel model.LikeCountModel
	BizRedis       *redis.Redis
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
		BizRedis:       rds,
		LikeModel:      model.NewLikeRecordModel(sqlx.NewMysql(c.DataSource)),
		LikeCountModel: model.NewLikeCountModel(sqlx.NewMysql(c.DataSource)),
	}
}
