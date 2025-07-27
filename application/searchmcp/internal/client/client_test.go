package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"posta/application/searchmcp/internal/config"

	"github.com/zeromicro/go-zero/core/conf"
)

func setupTestClient(t *testing.T) *ESClient {
	// 加载配置
	var c config.Config
	conf.MustLoad("../../etc/searchmcp.yaml", &c)

	// 创建客户端
	client, err := NewESClient(c)
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	return client
}

func TestFullWorkflow(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	testIndex := fmt.Sprintf("test-index-%d", time.Now().Unix())
	testAlias := fmt.Sprintf("test-alias-%d", time.Now().Unix())

	// 清理函数
	defer func() {
		// 删除别名（如果存在）
		client.DeleteAlias(ctx, testIndex, testAlias)
		// 删除索引
		client.DeleteIndex(ctx, testIndex)
	}()

	t.Run("1. Test Cluster Health", func(t *testing.T) {
		health, err := client.GetClusterHealth(ctx)
		if err != nil {
			t.Fatalf("Failed to get cluster health: %v", err)
		}
		t.Logf("Cluster health: %+v", health)
	})

	t.Run("2. Test Cluster Stats", func(t *testing.T) {
		stats, err := client.GetClusterStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get cluster stats: %v", err)
		}
		t.Logf("Cluster stats keys: %+v", getKeys(stats))
	})

	t.Run("3. List Indices (before)", func(t *testing.T) {
		indices, err := client.ListIndices(ctx)
		if err != nil {
			t.Fatalf("Failed to list indices: %v", err)
		}
		t.Logf("Current indices count: %d", len(indices))
		if len(indices) > 0 {
			t.Logf("First index example: %+v", indices[0])
		}
	})

	t.Run("4. Create Index", func(t *testing.T) {
		indexBody := map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 0,
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
					},
					"content": map[string]interface{}{
						"type": "text",
					},
					"created_at": map[string]interface{}{
						"type": "date",
					},
				},
			},
		}

		result, err := client.CreateIndex(ctx, testIndex, indexBody)
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}
		t.Logf("Create index result: %+v", result)
	})

	t.Run("5. Get Index Info", func(t *testing.T) {
		info, err := client.GetIndex(ctx, testIndex)
		if err != nil {
			t.Fatalf("Failed to get index info: %v", err)
		}
		t.Logf("Index info keys: %+v", getKeys(info))
	})

	t.Run("6. Index Document with ID", func(t *testing.T) {
		doc := map[string]interface{}{
			"title":      "Test Document 1",
			"content":    "This is a test document for our Elasticsearch client",
			"created_at": time.Now().Format(time.RFC3339),
		}

		docID := "doc1"
		result, err := client.IndexDocument(ctx, testIndex, doc, &docID)
		if err != nil {
			t.Fatalf("Failed to index document: %v", err)
		}
		t.Logf("Index document result: %+v", result)
	})

	t.Run("7. Index Document without ID", func(t *testing.T) {
		doc := map[string]interface{}{
			"title":      "Test Document 2",
			"content":    "This is another test document",
			"created_at": time.Now().Format(time.RFC3339),
		}

		result, err := client.IndexDocument(ctx, testIndex, doc, nil)
		if err != nil {
			t.Fatalf("Failed to index document: %v", err)
		}
		t.Logf("Index document result: %+v", result)
	})

	// 等待索引刷新
	time.Sleep(1 * time.Second)

	t.Run("8. Get Document", func(t *testing.T) {
		doc, err := client.GetDocument(ctx, testIndex, "doc1")
		if err != nil {
			t.Fatalf("Failed to get document: %v", err)
		}
		t.Logf("Retrieved document: %+v", doc)
	})

	t.Run("9. Search Documents", func(t *testing.T) {
		searchBody := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"title": "Test",
				},
			},
		}

		result, err := client.SearchDocuments(ctx, testIndex, searchBody)
		if err != nil {
			t.Fatalf("Failed to search documents: %v", err)
		}
		t.Logf("Search result keys: %+v", getKeys(result))

		if hits, ok := result["hits"].(map[string]interface{}); ok {
			if total, ok := hits["total"].(map[string]interface{}); ok {
				t.Logf("Total hits: %+v", total["value"])
			}
		}
	})

	t.Run("10. Create Alias", func(t *testing.T) {
		aliasBody := map[string]interface{}{
			"filter": map[string]interface{}{
				"term": map[string]interface{}{
					"title": "Test",
				},
			},
		}

		result, err := client.PutAlias(ctx, testIndex, testAlias, aliasBody)
		if err != nil {
			t.Fatalf("Failed to create alias: %v", err)
		}
		t.Logf("Create alias result: %+v", result)
	})

	t.Run("11. Get Alias", func(t *testing.T) {
		aliases, err := client.GetAlias(ctx, testIndex)
		if err != nil {
			t.Fatalf("Failed to get aliases: %v", err)
		}
		t.Logf("Index aliases: %+v", aliases)
	})

	t.Run("12. List All Aliases", func(t *testing.T) {
		aliases, err := client.ListAliases(ctx)
		if err != nil {
			t.Fatalf("Failed to list aliases: %v", err)
		}
		t.Logf("All aliases count: %d", len(aliases))
		if len(aliases) > 0 {
			t.Logf("First alias example: %+v", aliases[0])
		}
	})

	t.Run("13. Delete Document", func(t *testing.T) {
		result, err := client.DeleteDocument(ctx, testIndex, "doc1")
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}
		t.Logf("Delete document result: %+v", result)
	})

	t.Run("14. Delete by Query", func(t *testing.T) {
		// 只删除第二个文档，避免与已删除的doc1冲突
		deleteQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": map[string]interface{}{
						"term": map[string]interface{}{
							"_id": "doc1", // 排除已删除的文档
						},
					},
					"must": map[string]interface{}{
						"match": map[string]interface{}{
							"title": "Test Document 2",
						},
					},
				},
			},
		}

		result, err := client.DeleteByQuery(ctx, testIndex, deleteQuery)
		if err != nil {
			t.Fatalf("Failed to delete by query: %v", err)
		}
		t.Logf("Delete by query result: %+v", result)
	})

	t.Run("15. Delete Alias", func(t *testing.T) {
		result, err := client.DeleteAlias(ctx, testIndex, testAlias)
		if err != nil {
			t.Fatalf("Failed to delete alias: %v", err)
		}
		t.Logf("Delete alias result: %+v", result)
	})

	t.Run("16. Delete Index", func(t *testing.T) {
		result, err := client.DeleteIndex(ctx, testIndex)
		if err != nil {
			t.Fatalf("Failed to delete index: %v", err)
		}
		t.Logf("Delete index result: %+v", result)
	})
}

// 辅助函数：获取map的所有键
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestIndividualFunctions(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("Cluster Health", func(t *testing.T) {
		result, err := client.GetClusterHealth(ctx)
		if err != nil {
			t.Errorf("GetClusterHealth failed: %v", err)
		} else {
			t.Logf("Cluster health: %+v", result)
		}
	})

	t.Run("List Indices", func(t *testing.T) {
		result, err := client.ListIndices(ctx)
		if err != nil {
			t.Errorf("ListIndices failed: %v", err)
		} else {
			t.Logf("Indices count: %d", len(result))
		}
	})

	t.Run("List Aliases", func(t *testing.T) {
		result, err := client.ListAliases(ctx)
		if err != nil {
			t.Errorf("ListAliases failed: %v", err)
		} else {
			t.Logf("Aliases count: %d", len(result))
		}
	})
}

// 辅助函数：处理不同类型的结果
func logResult(t *testing.T, name string, result interface{}) {
	switch v := result.(type) {
	case []interface{}:
		t.Logf("%s returned array with %d items", name, len(v))
		if len(v) > 0 {
			t.Logf("First item: %+v", v[0])
		}
	case map[string]interface{}:
		t.Logf("%s returned object with keys: %+v", name, getKeys(v))
	default:
		t.Logf("%s returned type %T: %+v", name, v, v)
	}
}
