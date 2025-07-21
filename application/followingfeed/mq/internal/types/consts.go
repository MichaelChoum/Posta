package types

const (
	BigUpFansThreshold = 50000
	BatchSize          = 1000 //小批量插入的数目
)

const (
	FollowStatusFollow   = iota + 1 // 关注
	FollowStatusUnfollow            // 取消关注
)

const (
	// ArticleStatusPending 待审核
	ArticleStatusPending = iota
	// ArticleStatusNotPass 审核不通过
	ArticleStatusNotPass
	// ArticleStatusVisible 可见
	ArticleStatusVisible
	// ArticleStatusUserDelete 用户删除
	ArticleStatusUserDelete
)
