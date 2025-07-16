package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ LikeRecordModel = (*customLikeRecordModel)(nil)

type (
	// LikeRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customLikeRecordModel.
	LikeRecordModel interface {
		likeRecordModel
		withSession(session sqlx.Session) LikeRecordModel
		InsertIgnore(ctx context.Context, bizId, objId, userId int64) (sql.Result, error)
		DeleteByBizObjUser(ctx context.Context, bizId, objId, userId int64) error
	}

	customLikeRecordModel struct {
		*defaultLikeRecordModel
	}
)

// NewLikeRecordModel returns a model for the database table.
func NewLikeRecordModel(conn sqlx.SqlConn) LikeRecordModel {
	return &customLikeRecordModel{
		defaultLikeRecordModel: newLikeRecordModel(conn),
	}
}

func (m *customLikeRecordModel) withSession(session sqlx.Session) LikeRecordModel {
	return NewLikeRecordModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customLikeRecordModel) InsertIgnore(ctx context.Context, bizId, objId, userId int64) (sql.Result, error) {
	query := fmt.Sprintf("INSERT IGNORE INTO" + m.table + " (biz_id, obj_id, user_id) VALUES (?, ?, ?)")
	return m.conn.ExecCtx(ctx, query, bizId, objId, userId)
}

func (m *customLikeRecordModel) DeleteByBizObjUser(ctx context.Context, bizId, objId, userId int64) error {
	query := fmt.Sprintf("DELETE FROM" + m.table + "WHERE biz_id = ? AND obj_id = ? AND user_id = ?")
	_, err := m.conn.ExecCtx(ctx, query, bizId, objId, userId)
	return err
}
