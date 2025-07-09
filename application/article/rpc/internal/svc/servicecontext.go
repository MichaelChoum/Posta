package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/singleflight"
	"posta/application/article/rpc/internal/config"
	"posta/application/article/rpc/internal/model"
)

type ServiceContext struct {
	Config            config.Config
	ArticleModel      model.ArticleModel
	BizRedis          *redis.Redis
	SingleFlightGroup singleflight.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds, _ := redis.NewRedis(redis.RedisConf{
		Host:     c.BizRedis.Host,
		Pass:     c.BizRedis.Pass,
		Type:     c.BizRedis.Type,
		NonBlock: true, // 注意：避免因为超时就直接panic
	})

	return &ServiceContext{
		Config:       c,
		ArticleModel: model.NewArticleModel(sqlx.NewMysql(c.DataSource), c.CacheRedis),
		BizRedis:     rds,
	}
}
