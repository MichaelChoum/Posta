package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"posta/application/followingfeed/rpc/internal/types"
)

var _ FollowModel = (*customFollowModel)(nil)

type (
	// FollowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFollowModel.
	FollowModel interface {
		followModel
		withSession(session sqlx.Session) FollowModel
		GetFollowedIds(ctx context.Context, userId int64) ([]int64, error)
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

func (m *customFollowModel) GetFollowedIds(ctx context.Context, userId int64) ([]int64, error) {
	query := fmt.Sprintf("select followed_user_id from " + m.table + " where user_id = ? and follow_status = ?")
	var followedIds []int64
	err := m.conn.QueryRowsCtx(ctx, &followedIds, query, userId, types.FollowStatusFollow)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return followedIds, nil
}
