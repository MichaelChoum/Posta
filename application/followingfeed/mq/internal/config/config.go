package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Config struct {
	service.ServiceConf
	ArticleKqConsumerConf   kq.KqConf
	FollowKqConsumerConf    kq.KqConf
	DataSourceArticle       string
	DataSourceFollow        string
	DataSourceFollowingFeed string
	CacheRedis              cache.CacheConf
	BizRedis                redis.RedisConf
}
