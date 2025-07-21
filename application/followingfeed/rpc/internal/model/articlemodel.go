package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"posta/application/followingfeed/rpc/internal/types"
	"strings"
)

var _ ArticleModel = (*customArticleModel)(nil)

type (
	// ArticleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleModel.
	ArticleModel interface {
		articleModel
		ArticlesLiteByUserIds(ctx context.Context, userIds []int64, inCursor int64, pageSize int64) ([]types.ArticleLite, error)
	}

	customArticleModel struct {
		*defaultArticleModel
	}
)

// NewArticleModel returns a model for the database table.
func NewArticleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ArticleModel {
	return &customArticleModel{
		defaultArticleModel: newArticleModel(conn, c, opts...),
	}
}

func (m *customArticleModel) ArticlesLiteByUserIds(ctx context.Context, userIds []int64, inCursor int64, pageSize int64) ([]types.ArticleLite, error) {
	// 定义结果切片
	var articlesLite []types.ArticleLite

	// 检查输入参数
	if len(userIds) == 0 || pageSize <= 0 {
		return nil, fmt.Errorf("invalid input: userIds or pageSize must be greater than 0")
	}

	// 动态拼接 SQL
	query := `SELECT id, UNIX_TIMESTAMP(publish_time) AS publish_time
			  FROM article
			  WHERE publish_time < FROM_UNIXTIME(?) 
			  AND status = 2
			  AND author_id IN (` // 开始构造 IN 子句
	placeholders := make([]string, len(userIds))
	args := make([]interface{}, 0, len(userIds)+2)
	args = append(args, inCursor) // 第一个参数是分页游标 inCursor

	// 动态生成 IN 子句占位符和参数
	for i, id := range userIds {
		placeholders[i] = "?"   // 为每个 userId 插入一个占位符
		args = append(args, id) // 添加实际的 userId 参数
	}
	query += strings.Join(placeholders, ",") + `)
			  ORDER BY publish_time DESC
			  LIMIT ?`

	// 添加分页大小参数
	args = append(args, pageSize)

	// 使用 QueryRowsPartialCtx 实现逐行查询
	err := m.QueryRowsPartialNoCacheCtx(ctx, &articlesLite, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return articlesLite, nil
}
