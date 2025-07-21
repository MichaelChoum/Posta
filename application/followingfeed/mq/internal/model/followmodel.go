package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"posta/application/followingfeed/mq/internal/types"
)

var _ FollowModel = (*customFollowModel)(nil)

type (
	// FollowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFollowModel.
	FollowModel interface {
		followModel
		withSession(session sqlx.Session) FollowModel
		GetFanIds(ctx context.Context, authorId int64) ([]int64, error)
	}

	customFollowModel struct {
		*defaultFollowModel
	}
)

// NewFollowModel returns a model for the database table.
func NewFollowModel(conn sqlx.SqlConn) FollowModel {
	return &customFollowModel{
		defaultFollowModel: newFollowModel(conn),
	}
}

func (m *customFollowModel) withSession(session sqlx.Session) FollowModel {
	return NewFollowModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFollowModel) GetFanIds(ctx context.Context, authorId int64) ([]int64, error) {
	query := fmt.Sprintf("select user_id from " + m.table + " where followed_user_id = ? and follow_status = ?")
	var fanIds []int64
	err := m.conn.QueryRowsCtx(ctx, &fanIds, query, authorId, types.FollowStatusFollow)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return fanIds, nil
}
