package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/singleflight"
	"posta/application/followingfeed/rpc/internal/config"
	"posta/application/followingfeed/rpc/internal/model"
)

type ServiceContext struct {
	Config            config.Config
	UserInBoxModel    model.UserInboxModel
	FollowModel       model.FollowModel
	ArticleModel      model.ArticleModel
	FollowCountModel  model.FollowCountModel
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
		Config:           c,
		UserInBoxModel:   model.NewUserInboxModel(sqlx.NewMysql(c.DataSourceFollowingFeed)),
		FollowModel:      model.NewFollowModel(sqlx.NewMysql(c.DataSourceFollow)),
		FollowCountModel: model.NewFollowCountModel(sqlx.NewMysql(c.DataSourceFollow)),
		ArticleModel:     model.NewArticleModel(sqlx.NewMysql(c.DataSourceArticle), c.CacheRedis),
		BizRedis:         rds,
	}
}
