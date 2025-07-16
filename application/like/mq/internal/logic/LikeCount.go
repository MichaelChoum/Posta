package logic

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"posta/application/like/mq/internal/svc"
	"strconv"
	"strings"
	"time"
)

type LikeCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCountLogic {
	return &LikeCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

const dirtyKeysSetKey = "like:dirty_keys" // 存储redis中有变化的点赞数key名

func (l *LikeCountLogic) StartLikeCountFlusher(ctx context.Context) {
	threading.GoSafe(func() {
		// 聚合10s内的点赞数目
		// 使用 time.NewTicker 创建一个定时器（ticker），它会每隔 10 秒 发出一次信号（写入到它的 channel ticker.C 中）。
		// 定时器的声明和结束要写在里面，不然生命周期直接结束了
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()
				err := l.flushLikeCounts(ctx)
				if err != nil {
					l.Logger.Errorf("flush like counts error, %v", err)
					continue
				}
			}
		}
	})
}

func (l *LikeCountLogic) flushLikeCounts(ctx context.Context) error {
	dirtyKeys, err := l.svcCtx.BizRedis.SmembersCtx(ctx, dirtyKeysSetKey)
	if err != nil {
		l.Logger.Errorf("smembers_ctx err:%v", err)
		return err
	}
	for _, key := range dirtyKeys {
		bizId, objId, err := l.parseLikeCountKey(key)
		if err != nil {
			l.Logger.Errorf("parseLikeCountKey err:%v", err)
			continue
		}

		countStr, err := l.svcCtx.BizRedis.GetCtx(ctx, key)
		if err != nil || countStr == "" {
			l.Logger.Errorf("[Flusher] redis GET %s error: %v", key, err)
			continue
		}

		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			l.Logger.Errorf("[Flusher] parse count error for %s: %v", key, err)
			continue
		}

		err = l.svcCtx.LikeCountModel.InsertOrUpdateCount(ctx, bizId, objId, count)
		if err != nil {
			l.Logger.Errorf("[Flusher] db write error for %s: %v", key, err)
			continue
		}

		// 落库成功后移除脏 key
		_, _ = l.svcCtx.BizRedis.SremCtx(ctx, dirtyKeysSetKey, key)
	}
	return nil
}

func (l *LikeCountLogic) parseLikeCountKey(key string) (bizId, targetId int64, err error) {
	// 格式："biz#like_count#[bizid]#[targetid]"
	parts := strings.Split(key, "#")
	if len(parts) != 4 {
		err := fmt.Errorf("invalid likeCountKey format: %s", key)
		return 0, 0, err
	}
	bizId, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	targetId, err = strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return bizId, targetId, nil
}
