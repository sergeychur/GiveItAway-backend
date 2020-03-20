package models

type Comment struct {

	CommentId int32 `json:"comment_id,omitempty"`

	AuthorId int32 `json:"author_id,omitempty"`

	Text string `json:"text,omitempty"`
}