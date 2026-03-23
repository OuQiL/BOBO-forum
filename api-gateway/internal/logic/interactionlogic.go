package logic

import (
	"context"

	"api-gateway/internal/svc"
	"api-gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InteractionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInteractionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InteractionLogic {
	return &InteractionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InteractionLogic) Like(req *types.LikeRequest) (resp interface{}, err error) {
	// TODO: 调用interaction.rpc服务进行点赞
	// TODO: 检查是否已点赞（幂等性）
	// TODO: 更新帖子点赞计数
	// TODO: 使用Redis缓存点赞状态

	return map[string]interface{}{
		"message": "点赞成功",
		"post_id": req.PostID,
	}, nil
}

func (l *InteractionLogic) Unlike(req *types.LikeRequest) (resp interface{}, err error) {
	// TODO: 取消点赞实现

	return map[string]interface{}{
		"message": "取消点赞成功",
		"post_id": req.PostID,
	}, nil
}

func (l *InteractionLogic) Comment(req *types.CommentRequest) (resp interface{}, err error) {
	// TODO: 调用interaction.rpc服务进行评论
	// TODO: 评论内容敏感词过滤
	// TODO: 评论通知帖子作者

	return map[string]interface{}{
		"message": "评论成功",
		"post_id": req.PostID,
	}, nil
}

func (l *InteractionLogic) Follow(req *types.FollowRequest) (resp interface{}, err error) {
	// TODO: 调用interaction.rpc服务进行关注
	// TODO: 检查是否已关注（幂等性）
	// TODO: 更新用户粉丝计数
	// TODO: 关注通知被关注用户

	return map[string]interface{}{
		"message": "关注成功",
		"user_id": req.UserID,
	}, nil
}

func (l *InteractionLogic) Unfollow(req *types.FollowRequest) (resp interface{}, err error) {
	// TODO: 取消关注实现

	return map[string]interface{}{
		"message": "取消关注成功",
		"user_id": req.UserID,
	}, nil
}

func (l *InteractionLogic) GetFollowers(userID int64, page, size int) (resp interface{}, err error) {
	// TODO: 获取粉丝列表

	return map[string]interface{}{
		"total": 0,
		"list":  []interface{}{},
	}, nil
}

func (l *InteractionLogic) GetFollowing(userID int64, page, size int) (resp interface{}, err error) {
	// TODO: 获取关注列表

	return map[string]interface{}{
		"total": 0,
		"list":  []interface{}{},
	}, nil
}
