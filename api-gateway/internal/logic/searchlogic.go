package logic

import (
	"context"
	"fmt"

	"api-gateway/internal/svc"
	"api-gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
	pb "search/api/proto"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchLogic) Search(req *types.SearchRequest) (resp *types.SearchResponse, err error) {
	// 调用 search-rpc
	rpcReq := &pb.SearchRequest{
		Keyword: req.Keyword,
		Type:    req.Type,
		Page:    int32(req.Page),
		Size:    int32(req.Size),
		SortBy:  "relevance",
	}

	// 获取 RPC 客户端
	if l.svcCtx.SearchRpc == nil {
		return nil, fmt.Errorf("search rpc client not available")
	}

	client := pb.NewSearchClient(l.svcCtx.SearchRpc.Conn())
	rpcResp, err := client.Search(l.ctx, rpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return nil, fmt.Errorf("search rpc failed: %v", st.Message())
		}
		return nil, fmt.Errorf("search rpc failed: %w", err)
	}

	// 转换结果
	list := make([]interface{}, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		listItem := map[string]interface{}{
			"type":        item.Type,
			"id":          item.PostId,
			"user_id":     item.UserId,
			"username":    item.Username,
			"title":       item.Title,
			"content":     item.Content,
			"like_count":  item.LikeCount,
			"created_at":  item.CreatedAt,
			"highlight": map[string]string{
				"title":   item.HighlightTitle,
				"content": item.HighlightContent,
			},
		}
		list = append(list, listItem)
	}

	return &types.SearchResponse{
		Total: rpcResp.Total,
		List:  list,
	}, nil
}
