package main

import (
	"flag"
	"fmt"
	"posta/application/searchmcp/internal/client"
	"posta/application/searchmcp/internal/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/searchmcp.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化ES客户端
	esClient, err := client.NewESClient(c)
	if err != nil {
		logx.Errorf("Failed to initialize ES client: %v", err)
		return
	}

	logx.Info("Search MCP Server started successfully")

	// TODO: 这里后续会添加MCP服务器的启动逻辑
	fmt.Println("ES Client initialized successfully")

	// 简单测试连接
	info, err := esClient.GetClient().Info()
	if err != nil {
		logx.Errorf("Failed to get ES info: %v", err)
		return
	}
	defer info.Body.Close()

	fmt.Printf("Connected to Elasticsearch: %s\n", info.String())
}
