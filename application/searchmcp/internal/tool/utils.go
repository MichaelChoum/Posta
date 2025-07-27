package tool

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/mcp"
)

// FormatToolResult 将任意结果格式化为MCP可识别的内容
func FormatToolResult(result interface{}) (any, error) {
	if result == nil {
		return "No data returned", nil
	}

	switch v := result.(type) {
	case string:
		return v, nil
	case mcp.TextContent:
		return v, nil
	case mcp.ImageContent:
		return v, nil
	default:
		// 对于其他类型，序列化为JSON字符串
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("Failed to serialize result: %v", err), nil
		}
		return string(jsonBytes), nil
	}
}
