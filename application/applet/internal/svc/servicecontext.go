package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"posta/application/applet/internal/config"
	"posta/application/user/rpc/user"
)

type ServiceContext struct {
	Config   config.Config
	UserRPC  user.User
	BizRedis *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		UserRPC:  user.NewUser(zrpc.MustNewClient(c.UserRPC)),
		BizRedis: redis.New(c.BizRedis.Host, redis.WithPass(c.BizRedis.Pass)),
	}
}
