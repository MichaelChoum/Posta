package config

import (
	"github.com/zeromicro/go-zero/mcp"
)

type Config struct {
	//rest.RestConf
	Mcp mcp.McpConf
	Es  struct {
		Addresses []string
		Username  string
		Password  string
	}
}
