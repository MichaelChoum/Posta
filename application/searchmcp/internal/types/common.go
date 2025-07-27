package types

// ESResponse 通用ES响应结构
type ESResponse struct {
	Data   interface{} `json:"data"`
	Error  string      `json:"error,omitempty"`
	Status int         `json:"status"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
	Aliases  map[string]interface{} `json:"aliases,omitempty"`
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	Index   string                 `json:"_index"`
	ID      string                 `json:"_id"`
	Source  map[string]interface{} `json:"_source"`
	Version int64                  `json:"_version,omitempty"`
}
