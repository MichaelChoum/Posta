package types

type InboxItemLite struct {
	Id          int64
	ArticleId   int64
	PublishTime int64
}

type ArticleLite struct {
	ArticleId   int64 `db:"id"`
	PublishTime int64 `db:"publish_time"`
}
