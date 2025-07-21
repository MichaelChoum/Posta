package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserInboxModel = (*customUserInboxModel)(nil)

type (
	// UserInboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserInboxModel.
	UserInboxModel interface {
		userInboxModel
		withSession(session sqlx.Session) UserInboxModel
		UserInboxsByUserId(ctx context.Context, userId int64, pubTime string, limit int) ([]*UserInbox, error)
	}

	customUserInboxModel struct {
		*defaultUserInboxModel
	}
)

// NewUserInboxModel returns a model for the database table.
func NewUserInboxModel(conn sqlx.SqlConn) UserInboxModel {
	return &customUserInboxModel{
		defaultUserInboxModel: newUserInboxModel(conn),
	}
}

func (m *customUserInboxModel) withSession(session sqlx.Session) UserInboxModel {
	return NewUserInboxModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserInboxModel) UserInboxsByUserId(ctx context.Context, userId int64, pubTime string, limit int) ([]*UserInbox, error) {

	var userInboxs []*UserInbox
	sql := fmt.Sprintf("select " + userInboxRows + " from " + m.table + " where user_id=? and publish_time < ? and status=2 order by publish_time desc limit ?")
	err := m.conn.QueryRowsCtx(ctx, &userInboxs, sql, userId, pubTime, limit)
	if err != nil {
		return nil, err
	}
	return userInboxs, nil
}
