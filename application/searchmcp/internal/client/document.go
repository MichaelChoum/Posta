package client

import (
	"context"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// SearchDocuments 搜索文档
func (c *ESClient) SearchDocuments(ctx context.Context, index string, body map[string]interface{}) (map[string]interface{}, error) {
	var bodyReader *strings.Reader
	if body != nil {
		bodyBytes, err := jsonMarshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bodyReader,
	}
	return c.executeRequest(ctx, req)
}

// IndexDocument 索引文档（创建或更新）
func (c *ESClient) IndexDocument(ctx context.Context, index string, document map[string]interface{}, id *string) (map[string]interface{}, error) {
	bodyBytes, err := jsonMarshal(document)
	if err != nil {
		return nil, err
	}

	req := esapi.IndexRequest{
		Index:   index,
		Body:    strings.NewReader(string(bodyBytes)),
		Refresh: "true", // 立即刷新索引
	}

	if id != nil {
		req.DocumentID = *id
	}

	return c.executeRequest(ctx, req)
}

// GetDocument 获取文档
func (c *ESClient) GetDocument(ctx context.Context, index, id string) (map[string]interface{}, error) {
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}
	return c.executeRequest(ctx, req)
}

// DeleteDocument 删除文档
func (c *ESClient) DeleteDocument(ctx context.Context, index, id string) (map[string]interface{}, error) {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
	}
	return c.executeRequest(ctx, req)
}

// DeleteByQuery 根据查询删除文档
func (c *ESClient) DeleteByQuery(ctx context.Context, index string, body map[string]interface{}) (map[string]interface{}, error) {
	bodyBytes, err := jsonMarshal(body)
	if err != nil {
		return nil, err
	}

	req := esapi.DeleteByQueryRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(bodyBytes)),
	}
	return c.executeRequest(ctx, req)
}
