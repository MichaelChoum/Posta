package tool

import (
	"context"
	"fmt"
	"posta/application/searchmcp/internal/client"

	"github.com/zeromicro/go-zero/mcp"
)

// GetClusterHealthTool 获取集群健康状态的工具
func GetClusterHealthTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "get_cluster_health",
		Description: "Get basic information about the health of the Elasticsearch cluster",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "Limit the information returned to specific index (optional)",
				},
				"level": map[string]any{
					"type":        "string",
					"description": "Specify the level of detail for returned information (cluster, indices, shards)",
					"enum":        []string{"cluster", "indices", "shards"},
					"default":     "cluster",
				},
				"wait_for_status": map[string]any{
					"type":        "string",
					"description": "Wait until the cluster is at least the specified status (green, yellow, red)",
					"enum":        []string{"green", "yellow", "red"},
				},
				"timeout": map[string]any{
					"type":        "string",
					"description": "Explicit operation timeout",
					"default":     "30s",
				},
			},
			Required: []string{},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index         string `json:"index,optional"`
				Level         string `json:"level,optional"`
				WaitForStatus string `json:"wait_for_status,optional"`
				Timeout       string `json:"timeout,optional"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			// 注意：这里简化处理，实际的client.GetClusterHealth可能需要扩展以支持这些参数
			// 或者在这里使用GeneralAPIRequest来处理更复杂的参数
			health, err := esClient.GetClusterHealth(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster health: %w", err)
			}
			return FormatToolResult(health)
		},
	}
}

// GetClusterStatsTool 获取集群统计信息的工具
func GetClusterStatsTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "get_cluster_stats",
		Description: "Get high-level overview of cluster statistics including nodes, indices, and shards information",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"node_id": map[string]any{
					"type":        "string",
					"description": "Comma-separated list of node IDs or names to limit returned information (optional)",
				},
				"flat_settings": map[string]any{
					"type":        "boolean",
					"description": "Return settings in flat format",
					"default":     false,
				},
				"timeout": map[string]any{
					"type":        "string",
					"description": "Explicit operation timeout",
					"default":     "30s",
				},
			},
			Required: []string{},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				NodeID       string `json:"node_id,optional"`
				FlatSettings bool   `json:"flat_settings,optional"`
				Timeout      string `json:"timeout,optional"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			// 注意：这里简化处理，实际的client.GetClusterStats可能需要扩展以支持这些参数
			stats, err := esClient.GetClusterStats(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster stats: %w", err)
			}
			return FormatToolResult(stats)
		},
	}
}

// RegisterClusterTools 注册集群相关工具
func RegisterClusterTools(server mcp.McpServer, esClient *client.ESClient) {
	tools := []mcp.Tool{
		GetClusterHealthTool(esClient),
		GetClusterStatsTool(esClient),
	}

	for _, tool := range tools {
		server.RegisterTool(tool)
	}
}
