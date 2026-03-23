package types

type Request struct {
	Name string `path:"name,options=you|me"`
}

type Response struct {
	Message string `json:"message"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserInfo struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user_info"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PostResponse struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type SearchRequest struct {
	Keyword string `form:"keyword"`
	Type    string `form:"type,options=post|user"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=10"`
}

type SearchResponse struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

type LikeRequest struct {
	PostID int64 `json:"post_id"`
}

type CommentRequest struct {
	PostID  int64  `json:"post_id"`
	Content string `json:"content"`
}

type FollowRequest struct {
	UserID int64 `json:"user_id"`
}
