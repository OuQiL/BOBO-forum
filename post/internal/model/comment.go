package model

type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	Username  string
	Content   string
	ParentID  int64
	Status    int64
	CreatedAt int64
	UpdatedAt int64
}
