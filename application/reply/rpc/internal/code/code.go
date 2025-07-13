package code

import (
	"posta/pkg/xcode"
)

var (
	UserIdInvalid     = xcode.New(700001, "评论用户ID无效")
	ReplyContentEmpty = xcode.New(700002, "评论内容为空")
	ReplyIdInvalid    = xcode.New(700003, "评论ID无效")
	SortTypeInvalid   = xcode.New(700004, "评论排序类型无效")
	ArticleIdInvalid  = xcode.New(700005, "文章ID无效")
)
