package models

type Ad struct {

	AdId int64 `json:"ad_id,omitempty"`

	AuthorId int32 `json:"author_id,omitempty"`

	Header string `json:"header,omitempty"`

	Text string `json:"text,omitempty"`

	Region string `json:"region,omitempty"`

	District string `json:"district,omitempty"`

	IsAuction bool `json:"is_auction,omitempty"`

	FeedbackType string `json:"feedback_type,omitempty"`

	CreationDate string `json:"creation_date,omitempty"`

	MeetingPlace string `json:"meeting_place,omitempty"`

	Status string `json:"status,omitempty"`

	PathesToPhoto []string `json:"pathes_to_photo,omitempty"`

	Category string `json:"category,omitempty"`

	CommentsCount int32 `json:"comments_count,omitempty"`
}