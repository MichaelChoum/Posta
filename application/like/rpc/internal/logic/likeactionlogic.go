package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/threading"

	"posta/application/like/rpc/internal/svc"
	"posta/application/like/rpc/internal/types"
	"posta/application/like/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	prefixLikes     = "biz#like#%d#%d"       // bizid & userid -> set(targetid)
	prefixLikeCount = "biz#like_count#%d#%d" // bizid & targetid -> string(count)
	dirtyKeysSetKey = "like:dirty_keys"      // 存储redis中有变化的点赞数key名
	LikesExpire     = 3600 * 24
)

type LikeActionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeActionLogic {
	return &LikeActionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户对某个对象点赞或取消点赞
func (l *LikeActionLogic) LikeAction(in *pb.LikeActionRequest) (*pb.LikeActionResponse, error) {
	likeRecordKey := LikeRecordKey(in.BizId, in.UserId)
	likeCountKey := LikeCountKey(in.BizId, in.ObjId)

	var (
		ret *pb.LikeActionResponse
	)
	//// 检查是否点过赞，这里只查询redis即可。因为前端页面打开是查询过点赞记录并放到redis中了。
	//// Sismember: Set is member
	//exist, err := l.svcCtx.BizRedis.SismemberCtx(l.ctx, likeRecordKey, in.ObjId)
	//if err != nil {
	//	l.Logger.Error("redis %s Sismember error: %v", likeRecordKey, err)
	//	return nil, err
	//}

	// 如果是点赞操作
	if in.Action == 0 {

		// lua脚本保证原子性
		likeScript := `
		local exists = redis.call("SISMEMBER", KEYS[1], ARGV[1])
		if exists == 1 then
			return 0
		end
		redis.call("SADD", KEYS[1], ARGV[1])
		redis.call("INCR", KEYS[2])
		redis.call("SADD", KEYS[3], KEYS[2])
		redis.call("EXPIRE", KEYS[1], ARGV[2])
		redis.call("EXPIRE", KEYS[2], ARGV[3])
		return 1`

		res, err := l.svcCtx.BizRedis.EvalCtx(
			l.ctx,
			likeScript,
			[]string{
				likeRecordKey,
				likeCountKey,
				dirtyKeysSetKey,
			},
			in.ObjId,
			int(LikesExpire),   // record key 过期时间
			int(7*LikesExpire), // count key 过期时间
		)

		if err != nil {
			l.Logger.Errorf("lua EvalCtx error: %v", err)
			return nil, err
		}

		if res == int64(0) {
			// 已点赞，直接返回当前状态
			return &pb.LikeActionResponse{
				Success: true,
				Liked:   true,
			}, nil
		} else {
			ret = &pb.LikeActionResponse{
				Success: true,
				Liked:   true,
			}
		}

		//// 添加点赞记录
		//// _是成功加入的数量，如果已经存在，则返回0，不存在则返回1
		//_, err := l.svcCtx.BizRedis.SaddCtx(l.ctx, likeRecordKey, in.ObjId)
		//if err != nil {
		//	l.Logger.Error("redis %s Sadd error: %v", likeRecordKey, err)
		//	return nil, err
		//}
		//
		//// 点赞计数+1
		//// _是新的点赞总数
		//_, err = l.svcCtx.BizRedis.IncrCtx(l.ctx, likeCountKey)
		//if err != nil {
		//	l.Logger.Error("redis %s Incr error: %v", likeCountKey, err)
		//	return nil, err
		//}
		//_, err = l.svcCtx.BizRedis.SaddCtx(l.ctx, dirtyKeysSetKey, likeCountKey)
		//
		//_ = l.svcCtx.BizRedis.ExpireCtx(l.ctx, likeRecordKey, LikesExpire)
		//_ = l.svcCtx.BizRedis.ExpireCtx(l.ctx, likeCountKey, 7*LikesExpire)
		//ret = &pb.LikeActionResponse{
		//	Success: true,
		//	Liked:   true,
		//}
	} else {
		// 使用lua脚本保证原子性
		cancelScript := `
		local exists = redis.call("SISMEMBER", KEYS[1], ARGV[1])
		if exists == 0 then
			return 0
		end
		redis.call("SREM", KEYS[1], ARGV[1])
		redis.call("DECR", KEYS[2])
		redis.call("SADD", KEYS[3], KEYS[2])
		redis.call("EXPIRE", KEYS[1], ARGV[2])
		redis.call("EXPIRE", KEYS[2], ARGV[3])
		return 1
		`

		res, err := l.svcCtx.BizRedis.EvalCtx(
			l.ctx,
			cancelScript,
			[]string{
				likeRecordKey,
				likeCountKey,
				dirtyKeysSetKey,
			},
			in.ObjId,
			int(LikesExpire),
			int(7*LikesExpire),
		)

		if err != nil {
			l.Logger.Errorf("cancel like EvalCtx error: %v", err)
			return nil, err
		}

		if res == int64(0) {
			return &pb.LikeActionResponse{
				Success: true,
				Liked:   false,
			}, nil
		} else {
			ret = &pb.LikeActionResponse{
				Success: true,
				Liked:   false,
			}
		}

		//// 没有点赞过还要取消
		//if !exist {
		//	return &pb.LikeActionResponse{
		//		Success: true,
		//		Liked:   false,
		//	}, nil
		//}
		//
		//// _是成功移除的数量，如果确实存在并被移除，则返回1，否则返回0
		//// Srem: Set remove
		//_, err := l.svcCtx.BizRedis.SremCtx(l.ctx, likeRecordKey, in.ObjId)
		//if err != nil {
		//	l.Logger.Error("redis %s Srem error: %v", likeRecordKey, err)
		//	return nil, err
		//}
		//
		//_, err = l.svcCtx.BizRedis.DecrCtx(l.ctx, likeCountKey)
		//if err != nil {
		//	l.Logger.Error("redis %s Decr error: %v", likeCountKey, err)
		//	return nil, err
		//}
		//_, err = l.svcCtx.BizRedis.SaddCtx(l.ctx, dirtyKeysSetKey, likeCountKey)
		//
		//_ = l.svcCtx.BizRedis.ExpireCtx(l.ctx, likeRecordKey, LikesExpire)
		//_ = l.svcCtx.BizRedis.ExpireCtx(l.ctx, likeCountKey, 7*LikesExpire)
		//
		//ret = &pb.LikeActionResponse{
		//	Success: true,
		//	Liked:   false,
		//}
	}

	msg := &types.LikeActionMsg{
		BizId:      in.BizId,
		ObjId:      in.ObjId,
		UserId:     in.UserId,
		LikeAction: in.Action,
	}

	// 发送kafka消息，异步
	threading.GoSafe(func() {
		data, err := json.Marshal(msg)
		if err != nil {
			l.Logger.Errorf("LikeActionMsg marshal msg: %v error: %v", msg, err)
			return
		}
		ctx := context.Background()
		err = l.svcCtx.KqPusherClient.Push(ctx, string(data))
		if err != nil {
			l.Logger.Errorf("LikeActionMsg kq push data: %s error: %v", data, err)
		}
	})

	return ret, nil
}

func LikeRecordKey(bizId int64, userId int64) string {
	return fmt.Sprintf(prefixLikes, bizId, userId)
}

func LikeCountKey(bizId int64, targetId int64) string {
	return fmt.Sprintf(prefixLikeCount, bizId, targetId)
}
