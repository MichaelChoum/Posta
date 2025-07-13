package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/singleflight"
	"posta/application/reply/rpc/internal/config"
	"posta/application/reply/rpc/internal/model"
)

type ServiceContext struct {
	Config            config.Config
	ReplyModel        model.ReplyModel
	BizRedis          *redis.Redis
	SingleFlightGroup singleflight.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds, _ := redis.NewRedis(redis.RedisConf{
		Host:     c.BizRedis.Host,
		Pass:     c.BizRedis.Pass,
		Type:     c.BizRedis.Type,
		NonBlock: true,
	})

	return &ServiceContext{
		Config:     c,
		ReplyModel: model.NewReplyModel(sqlx.NewMysql(c.DataSource), c.CacheRedis),
		BizRedis:   rds,
	}
}
