package model

type Post struct {
	ID           int64
	UserID       int64
	CommunityID  int64
	Username     string
	Title        string
	Content      string
	Tags         string
	LikeCount    int64
	CommentCount int64
	ViewCount    int64
	CreatedAt    int64
	UpdatedAt    int64
}
