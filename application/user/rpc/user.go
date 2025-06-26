package main

import (
	"flag"
	"fmt"
	"posta/pkg/interceptors"

	"posta/application/user/rpc/internal/config"
	"posta/application/user/rpc/internal/server"
	"posta/application/user/rpc/internal/svc"
	"posta/application/user/rpc/service"

	"github.com/zeromicro/go-zero/core/conf"
	sv "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		service.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		// 注意：因为server模块重名了，这做了别名处理
		if c.Mode == sv.DevMode || c.Mode == sv.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 注意：自定义拦截器
	s.AddUnaryInterceptors(interceptors.ServerErrorInterceptor())

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
