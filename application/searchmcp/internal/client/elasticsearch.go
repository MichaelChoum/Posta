package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"posta/application/searchmcp/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/zeromicro/go-zero/core/logx"
)

// ESClient Elasticsearch客户端封装
type ESClient struct {
	client *elasticsearch.Client
	config config.Config
}

// NewESClient 创建新的ES客户端
func NewESClient(c config.Config) (*ESClient, error) {
	cfg := elasticsearch.Config{
		Addresses: c.Es.Addresses,
		Username:  c.Es.Username,
		Password:  c.Es.Password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logx.Errorf("Failed to create elasticsearch client: %v", err)
		return nil, err
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		logx.Errorf("Failed to connect to elasticsearch: %v", err)
		return nil, err
	}
	res.Body.Close()

	logx.Info("Successfully connected to Elasticsearch")

	return &ESClient{
		client: client,
		config: c,
	}, nil
}

// GetClient 获取原始ES客户端
func (c *ESClient) GetClient() *elasticsearch.Client {
	return c.client
}

// executeRequest 执行ES请求并返回响应
func (c *ESClient) executeRequest(ctx context.Context, req esapi.Request) (map[string]interface{}, error) {
	res, err := req.Do(ctx, c.client)
	if err != nil {
		logx.Errorf("ES request failed: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logx.Errorf("Failed to decode ES response: %v", err)
		return nil, err
	}

	if res.IsError() {
		logx.Errorf("ES returned error: %v", result)
	}

	return result, nil
}

// executeArrayRequest 执行返回数组的请求
func (c *ESClient) executeArrayRequest(ctx context.Context, req esapi.Request) ([]interface{}, error) {
	res, err := req.Do(ctx, c.client)
	if err != nil {
		logx.Errorf("Failed to execute ES request: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logx.Errorf("Failed to read ES response body: %v", err)
		return nil, err
	}

	if res.IsError() {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			logx.Errorf("ES returned error: %+v", errorResponse)
		}
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logx.Errorf("Failed to decode ES array response: %v", err)
		return nil, err
	}

	return result, nil
}

// jsonMarshal 辅助函数，用于序列化JSON
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// jsonUnmarshal 辅助函数，用于反序列化JSON
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
