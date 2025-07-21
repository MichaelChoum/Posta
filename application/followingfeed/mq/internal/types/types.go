package types

// CanalLikeMsg canal解析like binlog消息.
type CanalArticleMsg struct {
	Data []struct {
		ID          string `json:"id"`
		AuthorId    string `json:"author_id"`
		Status      string `json:"status"`
		PublishTime string `json:"publish_time"`
	}
}

type CanalFollowMsg struct {
	Data []struct {
		ID             string `json:"id"`
		UserId         string `json:"user_id"`
		FollowedUserID string `json:"followed_user_id"`
		Status         string `json:"follow_status"`
	}
}
