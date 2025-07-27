package tool

import (
	"context"
	"fmt"
	"posta/application/searchmcp/internal/client"

	"github.com/zeromicro/go-zero/mcp"
)

// ListIndicesTool 列出所有索引的工具
func ListIndicesTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "list_indices",
		Description: "List all indices in the Elasticsearch cluster",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{},
			Required:   []string{},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			indices, err := esClient.ListIndices(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list indices: %w", err)
			}
			return FormatToolResult(indices)
		},
	}
}

// GetIndexTool 获取索引信息的工具
func GetIndexTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "get_index",
		Description: "Get information about a specific index including mappings, settings, and aliases",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The name of the index to get information about",
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

			indexInfo, err := esClient.GetIndex(ctx, req.Index)
			if err != nil {
				return nil, fmt.Errorf("failed to get index %s: %w", req.Index, err)
			}
			return FormatToolResult(indexInfo)
		},
	}
}

// CreateIndexTool 创建索引的工具
func CreateIndexTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "create_index",
		Description: "Create a new index with optional settings and mappings",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The name of the index to create",
				},
				"settings": map[string]any{
					"type":        "object",
					"description": "Index settings (optional)",
				},
				"mappings": map[string]any{
					"type":        "object",
					"description": "Index mappings (optional)",
				},
			},
			Required: []string{"index"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index    string                 `json:"index"`
				Settings map[string]interface{} `json:"settings,optional"`
				Mappings map[string]interface{} `json:"mappings,optional"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			// 构建索引body
			body := make(map[string]interface{})
			if req.Settings != nil {
				body["settings"] = req.Settings
			}
			if req.Mappings != nil {
				body["mappings"] = req.Mappings
			}

			result, err := esClient.CreateIndex(ctx, req.Index, body)
			if err != nil {
				return nil, fmt.Errorf("failed to create index %s: %w", req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// DeleteIndexTool 删除索引的工具
func DeleteIndexTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "delete_index",
		Description: "Delete an index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The name of the index to delete",
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

			result, err := esClient.DeleteIndex(ctx, req.Index)
			if err != nil {
				return nil, fmt.Errorf("failed to delete index %s: %w", req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// RegisterIndexTools 注册索引相关工具
func RegisterIndexTools(server mcp.McpServer, esClient *client.ESClient) {
	tools := []mcp.Tool{
		ListIndicesTool(esClient),
		GetIndexTool(esClient),
		CreateIndexTool(esClient),
		DeleteIndexTool(esClient),
	}

	for _, tool := range tools {
		server.RegisterTool(tool)
	}
}
