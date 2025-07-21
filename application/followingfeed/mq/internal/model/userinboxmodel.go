package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"posta/application/followingfeed/mq/internal/types"
	"strings"
)

var _ UserInboxModel = (*customUserInboxModel)(nil)

type (
	// UserInboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserInboxModel.
	UserInboxModel interface {
		userInboxModel
		withSession(session sqlx.Session) UserInboxModel
		BatchInsert(ctx context.Context, data []*UserInbox) (sql.Result, error)
		BatchSetDeleteByArticle(ctx context.Context, articleId int64) (sql.Result, error)
		BatchSetDeleteBySenRec(ctx context.Context, userId, senderId int64) (sql.Result, error)
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

// 批量插入
func (m *customUserInboxModel) BatchInsert(ctx context.Context, data []*UserInbox) (sql.Result, error) {
	// 字段顺序要和 SQL里的values一致!
	if len(data) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf(
		"insert ignore into " + m.table + " (user_id, article_id, sender_id, publish_time, status, is_read) values ",
	)
	// 动态拼接values (?, ?, ...)(?, ?, ...)
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*6)
	for _, row := range data {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			row.UserId,
			row.ArticleId,
			row.SenderId,
			row.PublishTime,
			row.Status,
			row.IsRead,
		)
	}
	// 比如[]string{"(?, ?, ?, ?, ?, ?)", "(?, ?, ?, ?, ?, ?)", "(?, ?, ?, ?, ?, ?)"}
	// 用了join就会变成"(?, ?, ?, ?, ?, ?),(?, ?, ?, ?, ?, ?),(?, ?, ?, ?, ?, ?)"
	query += strings.Join(valueStrings, ",")
	return m.conn.ExecCtx(ctx, query, valueArgs...)
}

// 这里是软删除
func (m *defaultUserInboxModel) BatchSetDeleteByArticle(ctx context.Context, articleId int64) (sql.Result, error) {
	query := fmt.Sprintf("update " + m.table + " set status = ? where article_id = ?")
	return m.conn.ExecCtx(ctx, query, types.ArticleStatusUserDelete, articleId)
}

// 软删除: 根据 senderId 和 receiver（user）Id 更新 inbox
func (m *defaultUserInboxModel) BatchSetDeleteBySenRec(ctx context.Context, userId, senderId int64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set status=? where user_id=? and sender_id=?", m.table)
	return m.conn.ExecCtx(ctx, query, types.ArticleStatusUserDelete, userId, senderId)
}
