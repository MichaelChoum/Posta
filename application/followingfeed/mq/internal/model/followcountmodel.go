package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FollowCountModel = (*customFollowCountModel)(nil)

type (
	// FollowCountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFollowCountModel.
	FollowCountModel interface {
		followCountModel
		withSession(session sqlx.Session) FollowCountModel
		GetFansCount(ctx context.Context, userId int64) (int64, error)
	}

	customFollowCountModel struct {
		*defaultFollowCountModel
	}
)

// NewFollowCountModel returns a model for the database table.
func NewFollowCountModel(conn sqlx.SqlConn) FollowCountModel {
	return &customFollowCountModel{
		defaultFollowCountModel: newFollowCountModel(conn),
	}
}

func (m *customFollowCountModel) withSession(session sqlx.Session) FollowCountModel {
	return NewFollowCountModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFollowCountModel) GetFansCount(ctx context.Context, userId int64) (int64, error) {
	query := fmt.Sprintf("select fans_count from " + m.table + "where user_id = ?")
	var fansCount int64
	err := m.conn.QueryRowCtx(ctx, &fansCount, query, userId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return fansCount, nil
}
