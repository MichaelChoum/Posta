package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

var _ ReplyModel = (*customReplyModel)(nil)

type (
	// ReplyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customReplyModel.
	ReplyModel interface {
		replyModel
		UpdateReplyStatus(ctx context.Context, id int64, status int) error
		FirstRepliesByArticleId(ctx context.Context, articleId, likeNum int64, createTime, sortField string, limit int) ([]*Reply, error)
		SecondRepliesByFirstReplyId(ctx context.Context, parentId, likeNum int64, createTime, sortField string, limit int) ([]*Reply, error)
		MaxCreateTimeByParentId(ctx context.Context, parentId int64) (time.Time, error)
		RepliesByRootReplyId(ctx context.Context, rootReplyId int64, createTime string, limit int) ([]*Reply, error)
	}

	customReplyModel struct {
		*defaultReplyModel
	}
)

// NewReplyModel returns a model for the database table.
func NewReplyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ReplyModel {
	return &customReplyModel{
		defaultReplyModel: newReplyModel(conn, c, opts...),
	}
}

func (m *customReplyModel) UpdateReplyStatus(ctx context.Context, id int64, status int) error {
	postaReplyReplyIdKey := fmt.Sprintf("%s%v", cachePostaReplyReplyIdPrefix, id)
	// m.ExecCtx(...) 是一个封装好的方法：
	//传入上下文 ctx
	//给定一个函数体，在这个函数里执行 SQL 更新语句
	//同时传入缓存 key postaReplyReplyIdKey，通常意味着：如果更新成功，就清除这个 key 对应的缓存（缓存一致性机制）
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		// 注意：只能拼接表名，status 和 id 必须用参数绑定（即 ? 占位符），不能直接拼接进 SQL 字符串。
		// ❌ 错误示例：存在 SQL 注入风险
		//    userInput := "1; DROP TABLE reply --"
		//    query := fmt.Sprintf("UPDATE reply SET status = %d WHERE id = %s", status, userInput)
		//    结果执行：UPDATE reply SET status = 1 WHERE id = 1; DROP TABLE reply --
		//    会导致 reply 表被删除！
		//
		// ✅ 正确示例：参数绑定防止注入
		//    query := "UPDATE reply SET status = ? WHERE id = ?"
		//    conn.ExecCtx(ctx, query, status, userInput)
		//    即使 userInput 是恶意内容，也只会当成一个普通的字符串值处理，安全可靠。
		query := fmt.Sprintf("update %s set status = ? where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, status, id)
	}, postaReplyReplyIdKey)
	return err
}

func (m *customReplyModel) FirstRepliesByArticleId(ctx context.Context, articleId, likeNum int64, createTime, sortField string, limit int) ([]*Reply, error) {
	var (
		err error
		sql string
		// any可以接受任意类型的值
		anyField any
		replies  []*Reply
	)

	if sortField == "like_num" {
		anyField = likeNum
		// replyRows指的是字段名拼成的一个 SQL 用的字段列表，"`id`,`content`,`reply_user_id`,`target_id`,..."。
		sql = fmt.Sprintf("select "+replyRows+" from "+m.table+" where target_id=? and parent_id=0 and status=0 and like_num < ? order by %s desc limit ?", sortField)
	} else {
		anyField = createTime
		sql = fmt.Sprintf("select "+replyRows+" from "+m.table+" where target_id=? and parent_id=0 and status=0 and create_time < ? order by %s desc limit ?", sortField)
	}

	// 直接查数据库，因为我们已经在RepliesByArticleId外面查过缓存没找到了，这里就直接去数据库找就可以
	err = m.QueryRowsNoCacheCtx(ctx, &replies, sql, articleId, anyField, limit)

	if err != nil {
		return nil, err
	}
	return replies, nil
}

func (m *customReplyModel) SecondRepliesByFirstReplyId(ctx context.Context, parentId, likeNum int64, createTime, sortField string, limit int) ([]*Reply, error) {
	var (
		err error
		sql string
		// any可以接受任意类型的值
		anyField any
		replies  []*Reply
	)

	if sortField == "like_num" {
		anyField = likeNum
		// replyRows指的是字段名拼成的一个 SQL 用的字段列表，"`id`,`content`,`reply_user_id`,`target_id`,..."。
		sql = fmt.Sprintf("select "+replyRows+" from "+m.table+" where parent_id=? and status=0 and like_num < ? order by %s desc limit ?", sortField)
	} else {
		anyField = createTime
		// 二级评论查询要升序
		sql = fmt.Sprintf("select "+replyRows+" from "+m.table+" where parent_id=? and status=0 and create_time > ? order by %s asc limit ?", sortField)
	}

	// 直接查数据库，因为我们已经在RepliesByArticleId外面查过缓存没找到了，这里就直接去数据库找就可以
	err = m.QueryRowsNoCacheCtx(ctx, &replies, sql, parentId, anyField, limit)

	if err != nil {
		return nil, err
	}
	return replies, nil
}

// 注意：这个函数需要检验一下正不正确
func (m *customReplyModel) MaxCreateTimeByParentId(ctx context.Context, parentId int64) (time.Time, error) {
	var reply Reply
	logx.Infof("parentId: %d", parentId)
	query := fmt.Sprintf("SELECT * FROM %s WHERE parent_id = ? and status=0 ORDER BY create_time DESC LIMIT 1", m.table)

	err := m.QueryRowNoCacheCtx(ctx, &reply, query, parentId)
	if err != nil {
		return time.Time{}, err
	}
	return reply.CreateTime, nil
}

func (m *customReplyModel) RepliesByRootReplyId(ctx context.Context, rootReplyId int64, createTime string, limit int) ([]*Reply, error) {
	var replies []*Reply
	sql := fmt.Sprintf("select " + replyRows + " from " + m.table + " where root_reply_id=? and status=0 and create_time > ? order by create_time asc limit ?")
	err := m.QueryRowsNoCacheCtx(ctx, &replies, sql, rootReplyId, createTime, limit)
	if err != nil {
		return nil, err
	}
	return replies, nil
}
