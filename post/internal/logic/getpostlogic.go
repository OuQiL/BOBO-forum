package logic

import (
	"context"
	"encoding/json"
	"errors"

	"post/api/proto"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

var ErrPostNotFound = errors.New("post not found")

type GetPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostLogic {
	return &GetPostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPostLogic) GetPost(in *proto.GetPostRequest) (*proto.GetPostResponse, error) {
	if !l.svcCtx.BloomFilter.MightContain(in.PostId) {
		return nil, ErrPostNotFound
	}

	post, err := l.svcCtx.PostRepo.GetPostDetail(l.ctx, in.GetUserId(), in.GetPostId())
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	var tags []string
	if err := json.Unmarshal([]byte(post.Tags), &tags); err != nil {
		tags = []string{}
	}

	comments, err := l.svcCtx.CommentRepo.ListByPostID(l.ctx, in.PostId)
	if err != nil {
		comments = nil
	}

	commentInfos := make([]*proto.CommentInfo, 0, len(comments))
	for _, comment := range comments {
		commentInfos = append(commentInfos, &proto.CommentInfo{
			Id:        comment.ID,
			PostId:    comment.PostID,
			UserId:    comment.UserID,
			Username:  comment.Username,
			Content:   comment.Content,
			ParentId:  comment.ParentID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	return &proto.GetPostResponse{
		Post: &proto.PostInfo{
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
		},
		Comments: commentInfos,
	}, nil
}
