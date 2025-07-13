package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ReplySubCountModel = (*customReplySubCountModel)(nil)

type (
	// ReplySubCountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customReplySubCountModel.
	ReplySubCountModel interface {
		replySubCountModel
	}

	customReplySubCountModel struct {
		*defaultReplySubCountModel
	}
)

// NewReplySubCountModel returns a model for the database table.
func NewReplySubCountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ReplySubCountModel {
	return &customReplySubCountModel{
		defaultReplySubCountModel: newReplySubCountModel(conn, c, opts...),
	}
}
