package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GeneralAPIRequest 执行通用API请求
func (c *ESClient) GeneralAPIRequest(ctx context.Context, method, path string, params map[string]string, body map[string]interface{}) (interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := jsonMarshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	// 构建URL
	url := strings.TrimSuffix(c.config.Es.Addresses[0], "/") + "/" + strings.TrimPrefix(path, "/")

	// 添加查询参数
	if len(params) > 0 {
		values := make([]string, 0, len(params))
		for k, v := range params {
			values = append(values, fmt.Sprintf("%s=%s", k, v))
		}
		url += "?" + strings.Join(values, "&")
	}

	// 创建并执行请求
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, bodyReader)
	if err != nil {
		return nil, err
	}

	if c.config.Es.Username != "" {
		req.SetBasicAuth(c.config.Es.Username, c.config.Es.Password)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(respBody) == 0 {
		return map[string]interface{}{"status_code": resp.StatusCode}, nil
	}

	// 先尝试解析为对象
	var objResult map[string]interface{}
	if err := jsonUnmarshal(respBody, &objResult); err == nil {
		objResult["status_code"] = resp.StatusCode
		return objResult, nil
	}

	// 如果对象解析失败，尝试解析为数组
	var arrResult []interface{}
	if err := jsonUnmarshal(respBody, &arrResult); err == nil {
		return arrResult, nil
	}

	// 如果都失败了，返回原始字符串
	return map[string]interface{}{
		"raw_response": string(respBody),
		"status_code":  resp.StatusCode,
	}, nil
}
