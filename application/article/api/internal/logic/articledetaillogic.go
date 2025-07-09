package logic

import (
	"context"
	"fmt"
	"posta/application/article/rpc/article"
	"posta/application/user/rpc/user"
	"strconv"

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
	// 这里直接return是为了docker和k8s的快速测试。
	fmt.Print("1")
	return &types.ArticleDetailResponse{
		Title:       "文章标题1",
		Content:     "文章内容1",
		Description: "文章描述1",
		AuthorId:    "1",
		AuthorName:  "用户1",
	}, nil
	articleInfo, err := l.svcCtx.ArticleRPC.ArticleDetail(l.ctx, &article.ArticleDetailRequest{
		ArticleId: req.ArticleId,
	})
	if err != nil {
		logx.Errorf("get article detail id: %d err: %v", req.ArticleId, err)
		return nil, err
	}
	if articleInfo == nil || articleInfo.Article == nil {
		return nil, nil
	}
	userInfo, err := l.svcCtx.UserRPC.FindById(l.ctx, &user.FindByIdRequest{
		UserId: articleInfo.Article.AuthorId,
	})
	if err != nil {
		logx.Errorf("get userInfo id: %d err: %v", articleInfo.Article.AuthorId, err)
		return nil, err
	}

	return &types.ArticleDetailResponse{
		Title:       articleInfo.Article.Title,
		Content:     articleInfo.Article.Content,
		Description: articleInfo.Article.Description,
		Cover:       articleInfo.Article.Cover,
		AuthorId:    strconv.FormatInt(articleInfo.Article.AuthorId, 10),
		AuthorName:  userInfo.Username,
	}, nil
}
