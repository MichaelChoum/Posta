package client

import (
	"context"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ListAliases 获取所有别名
func (c *ESClient) ListAliases(ctx context.Context) ([]interface{}, error) {
	req := esapi.CatAliasesRequest{
		Format: "json",
	}
	return c.executeArrayRequest(ctx, req)
}

// GetAlias 获取指定索引的别名
func (c *ESClient) GetAlias(ctx context.Context, index string) (map[string]interface{}, error) {
	req := esapi.IndicesGetAliasRequest{
		Index: []string{index},
	}
	return c.executeRequest(ctx, req)
}

// PutAlias 创建或更新别名
func (c *ESClient) PutAlias(ctx context.Context, index, name string, body map[string]interface{}) (map[string]interface{}, error) {
	var bodyReader *strings.Reader
	if body != nil {
		bodyBytes, err := jsonMarshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	req := esapi.IndicesPutAliasRequest{
		Index: []string{index},
		Name:  name,
		Body:  bodyReader,
	}
	return c.executeRequest(ctx, req)
}

// DeleteAlias 删除别名
func (c *ESClient) DeleteAlias(ctx context.Context, index, name string) (map[string]interface{}, error) {
	req := esapi.IndicesDeleteAliasRequest{
		Index: []string{index},
		Name:  []string{name},
	}
	return c.executeRequest(ctx, req)
}
