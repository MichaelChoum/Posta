package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"posta/application/followingfeed/mq/internal/config"
	"posta/application/followingfeed/mq/internal/model"
)

type ServiceContext struct {
	Config           config.Config
	UserInBoxModel   model.UserInboxModel
	FollowCountModel model.FollowCountModel
	FollowModel      model.FollowModel
	ArticleModel     model.ArticleModel
	BizRedis         *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds, err := redis.NewRedis(redis.RedisConf{
		Host: c.BizRedis.Host,
		Pass: c.BizRedis.Pass,
		Type: c.BizRedis.Type,
	})
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:           c,
		UserInBoxModel:   model.NewUserInboxModel(sqlx.NewMysql(c.DataSourceFollowingFeed)),
		FollowModel:      model.NewFollowModel(sqlx.NewMysql(c.DataSourceFollow)),
		FollowCountModel: model.NewFollowCountModel(sqlx.NewMysql(c.DataSourceFollow)),
		ArticleModel:     model.NewArticleModel(sqlx.NewMysql(c.DataSourceArticle), c.CacheRedis),
		BizRedis:         rds,
	}
}
