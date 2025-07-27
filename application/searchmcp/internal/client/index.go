package client

import (
	"context"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ListIndices 列出所有索引
func (c *ESClient) ListIndices(ctx context.Context) ([]interface{}, error) {
	req := esapi.CatIndicesRequest{
		Format: "json",
	}
	return c.executeArrayRequest(ctx, req)
}

// GetIndex 获取索引信息
func (c *ESClient) GetIndex(ctx context.Context, index string) (map[string]interface{}, error) {
	req := esapi.IndicesGetRequest{
		Index: []string{index},
	}
	return c.executeRequest(ctx, req)
}

// CreateIndex 创建索引
func (c *ESClient) CreateIndex(ctx context.Context, index string, body map[string]interface{}) (map[string]interface{}, error) {
	var bodyReader *strings.Reader
	if body != nil {
		bodyBytes, err := jsonMarshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bodyReader,
	}
	return c.executeRequest(ctx, req)
}

// DeleteIndex 删除索引
func (c *ESClient) DeleteIndex(ctx context.Context, index string) (map[string]interface{}, error) {
	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}
	return c.executeRequest(ctx, req)
}
