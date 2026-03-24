package logic

import (
	"context"
	"encoding/json"

	"post/api/proto"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPostsLogic {
	return &ListPostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPostsLogic) ListPosts(in *proto.ListPostsRequest) (*proto.ListPostsResponse, error) {
	posts, total, err := l.svcCtx.PostRepo.List(l.ctx, in.Page, in.PageSize, in.CommunityId, in.Tag)
	if err != nil {
		return nil, err
	}

	postInfos := make([]*proto.PostInfo, 0, len(posts))
	for _, post := range posts {
		var tags []string
		if err := json.Unmarshal([]byte(post.Tags), &tags); err != nil {
			return nil, err
		}

		postInfos = append(postInfos, &proto.PostInfo{
			Id:           post.ID,
			UserId:       post.UserID,
			CommunityId:  post.CommunityID,
			Username:     post.Username,
			Title:        post.Title,
			Content:      post.Content,
			Tags:         tags,
			LikeCount:    post.LikeCount,
			CommentCount: post.CommentCount,
			ViewCount:    post.ViewCount,
			CreatedAt:    post.CreatedAt,
			UpdatedAt:    post.UpdatedAt,
		})
	}

	return &proto.ListPostsResponse{
		Posts: postInfos,
		Total: total,
	}, nil
}
