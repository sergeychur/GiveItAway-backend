package models

type Ad struct {
	AdId int64 `json:"ad_id,omitempty"`

	AuthorId int32 `json:"author_id"`

	Header string `json:"header"`

	Text string `json:"text"`

	Region string `json:"region"`

	District string `json:"district,omitempty"`

	IsAuction bool `json:"is_auction"`

	FeedbackType string `json:"feedback_type"`

	ExtraField string `json:"extra_field"`

	CreationDate string `json:"creation_date,omitempty"`

	GeoPosition *GeoPosition `json:"geo_position,omitempty"`

	Status string `json:"status,omitempty"`

	PathesToPhoto []string `json:"pathes_to_photo,omitempty"`

	Category string `json:"category"`

	CommentsCount int32 `json:"comments_count,omitempty"`
}

type AdCreationResult struct {
	AdId int64 `json:"ad_id,omitempty"`
}
