package logic

import (
	"context"

	"posta/application/article/api/internal/svc"
	"posta/application/article/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleDetailLogic {
	return &ArticleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleDetailLogic) ArticleDetail(req *types.ArticleDetailRequest) (resp *types.ArticleDetailResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
