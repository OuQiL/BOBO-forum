package handler

import (
	"net/http"

	"api-gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// TODO: 添加JWT认证中间件
	// TODO: 添加请求限流中间件
	// TODO: 添加链路追踪中间件
	// TODO: 添加请求日志中间件

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/register",
				Handler: RegisterHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/login",
				Handler: LoginHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/posts",
				Handler: CreatePostHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/v1/posts/:id",
				Handler: DeletePostHandler(serverCtx),
			},
			// TODO: 添加获取帖子详情接口 GET /api/v1/posts/:id
			// TODO: 添加获取帖子列表接口 GET /api/v1/posts
		},
		// TODO: 添加JWT认证 rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/search",
				Handler: SearchHandler(serverCtx),
			},
		},
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/interaction/like",
				Handler: LikeHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/interaction/comment",
				Handler: CommentHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/interaction/follow",
				Handler: FollowHandler(serverCtx),
			},
			// TODO: 添加取消点赞接口 DELETE /api/v1/interaction/like/:post_id
			// TODO: 添加取消关注接口 DELETE /api/v1/interaction/follow/:user_id
			// TODO: 添加获取粉丝列表接口 GET /api/v1/interaction/followers
			// TODO: 添加获取关注列表接口 GET /api/v1/interaction/following
		},
		// TODO: 添加JWT认证 rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)
}
