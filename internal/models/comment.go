package models

type Comment struct {
	CommentId int32 `json:"comment_id,omitempty"`

	AuthorId int32 `json:"author_id"`

	Text string `json:"text"`
}

type CommentForUser struct {
	CommentId int32 `json:"comment_id"`

	Author User `json:"author"`

	Text string `json:"text"`
}
