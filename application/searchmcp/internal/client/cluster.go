package client

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// GetClusterHealth 获取集群健康状态
func (c *ESClient) GetClusterHealth(ctx context.Context) (map[string]interface{}, error) {
	req := esapi.ClusterHealthRequest{}
	return c.executeRequest(ctx, req)
}

// GetClusterStats 获取集群统计信息
func (c *ESClient) GetClusterStats(ctx context.Context) (map[string]interface{}, error) {
	req := esapi.ClusterStatsRequest{}
	return c.executeRequest(ctx, req)
}
