package logic

import (
	"context"

	"api-gateway/internal/svc"
	"api-gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
	// TODO: 调用search.rpc服务进行搜索
	// TODO: 根据type区分搜索帖子还是用户
	// TODO: 实现全文搜索（可使用Elasticsearch）
	// TODO: 搜索结果高亮显示关键词
	// TODO: 搜索历史记录

	return &types.SearchResponse{
		Total: 0,
		List:  []interface{}{},
	}, nil
}

func (l *SearchLogic) SearchPosts(keyword string, page, size int) (resp *types.SearchResponse, err error) {
	// TODO: 搜索帖子实现
	// TODO: 支持标题和内容搜索
	// TODO: 搜索结果排序（按时间/热度）

	return &types.SearchResponse{
		Total: 0,
		List:  []interface{}{},
	}, nil
}

func (l *SearchLogic) SearchUsers(keyword string, page, size int) (resp *types.SearchResponse, err error) {
	// TODO: 搜索博主实现
	// TODO: 支持用户名和简介搜索

	return &types.SearchResponse{
		Total: 0,
		List:  []interface{}{},
	}, nil
}
