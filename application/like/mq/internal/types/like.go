package types

// 注意：这个types/like.go文件是自己手动创建的
type ThumbupMsg struct {
	BizId    string ` json:"bizId,omitempty"`    // 业务id
	ObjId    int64  ` json:"objId,omitempty"`    // 点赞对象id
	UserId   int64  ` json:"userId,omitempty"`   // 用户id
	LikeType int32  ` json:"likeType,omitempty"` // 类型
}
