package tool

import (
	"context"
	"fmt"
	"posta/application/searchmcp/internal/client"

	"github.com/zeromicro/go-zero/mcp"
)

// ListAliasesTool 列出所有别名的工具
func ListAliasesTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "list_aliases",
		Description: "List all aliases in the Elasticsearch cluster",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{},
			Required:   []string{},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			aliases, err := esClient.ListAliases(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list aliases: %w", err)
			}
			return FormatToolResult(aliases)
		},
	}
}

// GetAliasTool 获取指定索引别名的工具
func GetAliasTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "get_alias",
		Description: "Get alias information for a specific index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The name of the index to get aliases for",
				},
			},
			Required: []string{"index"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index string `json:"index"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			aliases, err := esClient.GetAlias(ctx, req.Index)
			if err != nil {
				return nil, fmt.Errorf("failed to get aliases for index %s: %w", req.Index, err)
			}
			return FormatToolResult(aliases)
		},
	}
}

// PutAliasTool 创建或更新别名的工具
func PutAliasTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "put_alias",
		Description: "Create or update an alias for a specific index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The name of the index to create an alias for",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "The name of the alias to create",
				},
				"body": map[string]any{
					"type":        "object",
					"description": "Optional alias configuration (filter, routing, etc.)",
				},
			},
			Required: []string{"index", "name"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index string                 `json:"index"`
				Name  string                 `json:"name"`
				Body  map[string]interface{} `json:"body,optional"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			result, err := esClient.PutAlias(ctx, req.Index, req.Name, req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to create alias %s for index %s: %w", req.Name, req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// DeleteAliasTool 删除别名的工具
func DeleteAliasTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "delete_alias",
		Description: "Delete an alias for a specific index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The name of the index to delete the alias from",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "The name of the alias to delete",
				},
			},
			Required: []string{"index", "name"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index string `json:"index"`
				Name  string `json:"name"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			result, err := esClient.DeleteAlias(ctx, req.Index, req.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to delete alias %s from index %s: %w", req.Name, req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// RegisterAliasTools 注册别名相关工具
func RegisterAliasTools(server mcp.McpServer, esClient *client.ESClient) {
	tools := []mcp.Tool{
		ListAliasesTool(esClient),
		GetAliasTool(esClient),
		PutAliasTool(esClient),
		DeleteAliasTool(esClient),
	}

	for _, tool := range tools {
		server.RegisterTool(tool)
	}
}
