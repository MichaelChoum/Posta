package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"posta/application/article/api/internal/logic"
	"posta/application/article/api/internal/svc"
)

func UploadCoverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUploadCoverLogic(r.Context(), svcCtx)
		// 注意：这里得加上这个r，因为UploadCoverLogic的UploadCover方法仍然需要使用到r.Request的上下文
		resp, err := l.UploadCover(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
