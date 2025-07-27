package tool

import (
	"posta/application/searchmcp/internal/client"

	"github.com/zeromicro/go-zero/mcp"
)

// RegisterAllTools 注册所有Elasticsearch工具
func RegisterAllTools(server mcp.McpServer, esClient *client.ESClient) {

	// 注册索引操作工具
	RegisterIndexTools(server, esClient)

	// 注册文档操作工具
	RegisterDocumentTools(server, esClient)

	// 注册集群操作工具
	RegisterClusterTools(server, esClient)

	// 注册别名操作工具
	RegisterAliasTools(server, esClient)
}
