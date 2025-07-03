package logic

import (
	"context"
	"posta/application/article/rpc/internal/code"
	"posta/application/article/rpc/internal/types"
	"posta/pkg/xcode"

	"posta/application/article/rpc/internal/svc"
	"posta/application/article/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleDeleteLogic {
	return &ArticleDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArticleDeleteLogic) ArticleDelete(in *pb.ArticleDeleteRequest) (*pb.ArticleDeleteResponse, error) {
	if in.UserId <= 0 {
		return nil, code.UserIdInvalid
	}
	if in.ArticleId <= 0 {
		return nil, code.ArticleIdInvalid
	}
	article, err := l.svcCtx.ArticleModel.FindOne(l.ctx, in.ArticleId)
	if err != nil {
		l.Logger.Errorf("ArticleDelete FindOne req: %v error: %v", in, err)
		return nil, err
	}
	// 注意：这里必须保证删除的文章必须为当前操作者所写。
	if article.AuthorId != in.UserId {
		return nil, xcode.AccessDenied
	}
	// 注意：删除文章要保证数据库和缓存的一致性。一般业务中并不真正在数据库中删除文章，而是改变文章的状态。
	err = l.svcCtx.ArticleModel.UpdateArticleStatus(l.ctx, in.ArticleId, types.ArticleStatusUserDelete)
	if err != nil {
		l.Logger.Errorf("UpdateArticleStatus req: %v error: %v", in, err)
		return nil, err
	}
	// 注意：删除文章要保证数据库和缓存的一致性。在上面操作完数据库之后，这里删除缓存中的数据。
	_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, articlesKey(in.UserId, types.SortPublishTime), in.ArticleId)
	if err != nil {
		l.Logger.Errorf("ZremCtx req: %v error: %v", in, err)
	}
	_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, articlesKey(in.UserId, types.SortLikeCount), in.ArticleId)
	if err != nil {
		l.Logger.Errorf("ZremCtx req: %v error: %v", in, err)
	}

	return &pb.ArticleDeleteResponse{}, nil
}
