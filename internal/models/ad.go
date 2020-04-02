package models

type Ad struct {
	AdId int64 `json:"ad_id,omitempty"`

	AuthorId int `json:"author_id"`

	Header string `json:"header"`

	Text string `json:"text"`

	Region string `json:"region"`

	District string `json:"district,omitempty"`

	IsAuction bool `json:"is_auction"`

	FeedbackType string `json:"feedback_type"`

	ExtraField string `json:"extra_field,omitempty"`

	CreationDate string `json:"creation_date,omitempty"`

	GeoPosition *GeoPosition `json:"geo_position,omitempty"`

	Status string `json:"status,omitempty"`

	PathesToPhoto []AdPhoto `json:"pathes_to_photo,omitempty"`

	Category string `json:"category"`

	CommentsCount int32 `json:"comments_count,omitempty"`
}

type AdCreationResult struct {
	AdId int64 `json:"ad_id,omitempty"`
}

type AdForUsers struct {
	AdId int64 `json:"ad_id,omitempty"`

	Author *User `json:"author,omitempty"`

	Header string `json:"header"`

	//Text string `json:"text"`

	Region string `json:"region"`

	District string `json:"district,omitempty"`

	IsAuction bool `json:"is_auction"`

	FeedbackType string `json:"feedback_type"`

	ExtraField string `json:"extra_field,omitempty"`

	CreationDate string `json:"creation_date,omitempty"`

	//GeoPosition *GeoPosition `json:"geo_position,omitempty"`

	Status string `json:"status,omitempty"`

	PathesToPhoto []AdPhoto `json:"pathes_to_photo,omitempty"`

	Category string `json:"category"`

	CommentsCount int32 `json:"comments_count,omitempty"`
}

type AdForUsersDetailed struct {
	AdId int64 `json:"ad_id,omitempty"`

	Author *User `json:"author,omitempty"`

	Header string `json:"header"`

	Text string `json:"text"`

	Region string `json:"region"`

	District string `json:"district,omitempty"`

	IsAuction bool `json:"is_auction"`

	FeedbackType string `json:"feedback_type"`

	ExtraField string `json:"extra_field,omitempty"`

	CreationDate string `json:"creation_date,omitempty"`

	GeoPosition *GeoPosition `json:"geo_position,omitempty"`

	Status string `json:"status,omitempty"`

	PathesToPhoto []AdPhoto `json:"pathes_to_photo,omitempty"`

	Category string `json:"category"`

	CommentsCount int32 `json:"comments_count,omitempty"`
	ViewsCount int32 `json:"views_count"`
}

type AdPhoto struct {
	AdPhotoId int
	PhotoUrl  string
}
