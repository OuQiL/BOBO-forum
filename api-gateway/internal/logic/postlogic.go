package logic

import (
	"context"

	"api-gateway/internal/svc"
	"api-gateway/internal/types"
	postproto "post/api/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type PostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostLogic {
	return &PostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PostLogic) CreatePost(req *types.CreatePostRequest) (resp *types.PostResponse, err error) {
	userID, ok := GetUserIDFromContext(l.ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	username, _ := GetUsernameFromContext(l.ctx)

	l.Infof("User %d (%s) creating post: %s", userID, username, req.Title)

	postResp, err := l.svcCtx.PostRpc.CreatePost(l.ctx, &postproto.CreatePostRequest{
		UserId:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}

	return &types.PostResponse{
		ID:      postResp.Post.Id,
		Title:   postResp.Post.Title,
		Content: postResp.Post.Content,
		Author:  username,
	}, nil
}

func (l *PostLogic) DeletePost(id int64) (resp interface{}, err error) {
	userID, ok := GetUserIDFromContext(l.ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	l.Infof("User %d deleting post: %d", userID, id)

	return map[string]interface{}{
		"message": "删除成功",
		"post_id": id,
	}, nil
}

func (l *PostLogic) GetPost(id int64) (resp *types.PostResponse, err error) {
	return &types.PostResponse{
		ID:      id,
		Title:   "示例帖子",
		Content: "这是帖子内容",
		Author:  "demo-user",
	}, nil
}

func (l *PostLogic) ListPosts(page, size int) (resp interface{}, err error) {
	return map[string]interface{}{
		"total": 100,
		"page":  page,
		"size":  size,
		"list":  []interface{}{},
	}, nil
}

func (l *PostLogic) GetPostDetail(id int64) (resp *types.GetPostDetailResponse, err error) {
	post, err := l.svcCtx.PostRpc.GetPost(l.ctx, &postproto.GetPostRequest{
		PostId: id,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetPostDetailResponse{
		ID:           post.Post.Id,
		UserID:       post.Post.UserId,
		CommunityID:  post.Post.CommunityId,
		Username:     post.Post.Username,
		Title:        post.Post.Title,
		Content:      post.Post.Content,
		Tags:         post.Post.Tags,
		LikeCount:    post.Post.LikeCount,
		CommentCount: post.Post.CommentCount,
		ViewCount:    post.Post.ViewCount,
		CreatedAt:    post.Post.CreatedAt,
		UpdatedAt:    post.Post.UpdatedAt,
	}, nil
}
