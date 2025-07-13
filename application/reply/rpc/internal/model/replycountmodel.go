package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ReplyCountModel = (*customReplyCountModel)(nil)

type (
	// ReplyCountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customReplyCountModel.
	ReplyCountModel interface {
		replyCountModel
	}

	customReplyCountModel struct {
		*defaultReplyCountModel
	}
)

// NewReplyCountModel returns a model for the database table.
func NewReplyCountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ReplyCountModel {
	return &customReplyCountModel{
		defaultReplyCountModel: newReplyCountModel(conn, c, opts...),
	}
}
