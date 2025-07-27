package main

import (
	"fmt"
	"log"
	"posta/application/searchmcp/internal/client"
	"posta/application/searchmcp/internal/config"
	"posta/application/searchmcp/internal/tool"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/mcp"
)

func main() {
	// 加载配置
	var c config.Config
	if err := conf.Load("etc/searchmcp.yaml", &c); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置日志
	logx.DisableStat()
	logx.Disable()

	// 创建Elasticsearch客户端
	esClient, err := client.NewESClient(c)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// 创建MCP服务器
	server := mcp.NewMcpServer(c.Mcp)
	defer server.Stop()

	// 注册所有工具
	tool.RegisterAllTools(server, esClient)

	// 启动服务器
	fmt.Printf("Starting Elasticsearch MCP server on %s:%d", c.Mcp.Host, c.Mcp.Port)
	server.Start()
}
