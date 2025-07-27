package tool

import (
	"context"
	"fmt"
	"posta/application/searchmcp/internal/client"

	"github.com/zeromicro/go-zero/mcp"
)

// SearchDocumentsTool 搜索文档的工具
func SearchDocumentsTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "search_documents",
		Description: "Search for documents in an index using Elasticsearch query DSL",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The index to search in",
				},
				"query": map[string]any{
					"type":        "object",
					"description": "Elasticsearch query DSL object",
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Number of results to return (default: 10)",
					"default":     10,
				},
				"from": map[string]any{
					"type":        "integer",
					"description": "Starting offset for results (default: 0)",
					"default":     0,
				},
			},
			Required: []string{"index", "query"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index string                 `json:"index"`
				Query map[string]interface{} `json:"query"`
				Size  int                    `json:"size,optional"`
				From  int                    `json:"from,optional"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			// 构建搜索body
			searchBody := map[string]interface{}{
				"query": req.Query,
			}
			if req.Size > 0 {
				searchBody["size"] = req.Size
			}
			if req.From > 0 {
				searchBody["from"] = req.From
			}

			result, err := esClient.SearchDocuments(ctx, req.Index, searchBody)
			if err != nil {
				return nil, fmt.Errorf("failed to search documents in index %s: %w", req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// IndexDocumentTool 索引文档的工具
func IndexDocumentTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "index_document",
		Description: "Create or update a document in an index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The index to store the document in",
				},
				"document": map[string]any{
					"type":        "object",
					"description": "The document to index",
				},
				"id": map[string]any{
					"type":        "string",
					"description": "Document ID (optional, will be auto-generated if not provided)",
				},
			},
			Required: []string{"index", "document"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index    string                 `json:"index"`
				Document map[string]interface{} `json:"document"`
				ID       string                 `json:"id,optional"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			var docID *string
			if req.ID != "" {
				docID = &req.ID
			}

			result, err := esClient.IndexDocument(ctx, req.Index, req.Document, docID)
			if err != nil {
				return nil, fmt.Errorf("failed to index document in index %s: %w", req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// GetDocumentTool 获取文档的工具
func GetDocumentTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "get_document",
		Description: "Get a document by ID from an index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The index containing the document",
				},
				"id": map[string]any{
					"type":        "string",
					"description": "The document ID",
				},
			},
			Required: []string{"index", "id"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index string `json:"index"`
				ID    string `json:"id"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			result, err := esClient.GetDocument(ctx, req.Index, req.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get document %s from index %s: %w", req.ID, req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// DeleteDocumentTool 删除文档的工具
func DeleteDocumentTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "delete_document",
		Description: "Delete a document by ID from an index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The index containing the document",
				},
				"id": map[string]any{
					"type":        "string",
					"description": "The document ID to delete",
				},
			},
			Required: []string{"index", "id"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index string `json:"index"`
				ID    string `json:"id"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			result, err := esClient.DeleteDocument(ctx, req.Index, req.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to delete document %s from index %s: %w", req.ID, req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// DeleteByQueryTool 根据查询删除文档的工具
func DeleteByQueryTool(esClient *client.ESClient) mcp.Tool {
	return mcp.Tool{
		Name:        "delete_by_query",
		Description: "Delete documents matching a query from an index",
		InputSchema: mcp.InputSchema{
			Properties: map[string]any{
				"index": map[string]any{
					"type":        "string",
					"description": "The index to delete documents from",
				},
				"query": map[string]any{
					"type":        "object",
					"description": "Elasticsearch query DSL object to match documents for deletion",
				},
			},
			Required: []string{"index", "query"},
		},
		Handler: func(ctx context.Context, params map[string]any) (any, error) {
			var req struct {
				Index string                 `json:"index"`
				Query map[string]interface{} `json:"query"`
			}

			if err := mcp.ParseArguments(params, &req); err != nil {
				return nil, fmt.Errorf("failed to parse params: %w", err)
			}

			result, err := esClient.DeleteByQuery(ctx, req.Index, req.Query)
			if err != nil {
				return nil, fmt.Errorf("failed to delete documents from index %s: %w", req.Index, err)
			}
			return FormatToolResult(result)
		},
	}
}

// RegisterDocumentTools 注册文档相关工具
func RegisterDocumentTools(server mcp.McpServer, esClient *client.ESClient) {
	tools := []mcp.Tool{
		SearchDocumentsTool(esClient),
		IndexDocumentTool(esClient),
		GetDocumentTool(esClient),
		DeleteDocumentTool(esClient),
		DeleteByQueryTool(esClient),
	}

	for _, tool := range tools {
		server.RegisterTool(tool)
	}
}
