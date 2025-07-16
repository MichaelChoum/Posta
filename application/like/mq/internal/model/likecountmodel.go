package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ LikeCountModel = (*customLikeCountModel)(nil)

type (
	// LikeCountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customLikeCountModel.
	LikeCountModel interface {
		likeCountModel
		withSession(session sqlx.Session) LikeCountModel
		InsertOrUpdateCount(ctx context.Context, bizId, objId int64, count int64) error
	}

	customLikeCountModel struct {
		*defaultLikeCountModel
	}
)

// NewLikeCountModel returns a model for the database table.
func NewLikeCountModel(conn sqlx.SqlConn) LikeCountModel {
	return &customLikeCountModel{
		defaultLikeCountModel: newLikeCountModel(conn),
	}
}

func (m *customLikeCountModel) withSession(session sqlx.Session) LikeCountModel {
	return NewLikeCountModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customLikeCountModel) InsertOrUpdateCount(ctx context.Context, bizId, objId int64, count int64) error {
	// 如果插入时发生唯一约束（如主键或唯一索引）冲突，则改为执行后面的更新语句
	query := fmt.Sprintf("INSERT INTO" + m.table + "(biz_id, obj_id, like_num) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE like_num = VALUES(like_num)")
	_, err := m.conn.ExecCtx(ctx, query, bizId, objId, count)
	return err
}
