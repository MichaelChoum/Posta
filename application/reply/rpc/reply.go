package main

import (
	"flag"
	"fmt"

	"posta/application/reply/rpc/internal/config"
	"posta/application/reply/rpc/internal/server"
	"posta/application/reply/rpc/internal/svc"
	"posta/application/reply/rpc/service"

	"github.com/zeromicro/go-zero/core/conf"
	sv "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/reply.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		service.RegisterReplyServer(grpcServer, server.NewReplyServer(ctx))

		if c.Mode == sv.DevMode || c.Mode == sv.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
