package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Es struct {
		Addresses []string
		Username  string
		Password  string
	}
}
