package models

type Ad struct {
	AdId int64 `json:"ad_id,omitempty"`

	AuthorId int `json:"author_id"`

	Header string `json:"header"`

	Text string `json:"text"`

	Region string `json:"region"`

	District string `json:"district,omitempty"`

	//IsAuction bool `json:"is_auction"`
	AdType string `json:"ad_type"`

	//FeedbackType string `json:"feedback_type"`
	LSEnabled bool `json:"ls_enabled"`

	CommentsEnabled bool `json:"comments_enabled"`

	ExtraEnabled bool `json:"extra_enabled"`

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

	AdType string `json:"ad_type"`

	//FeedbackType string `json:"feedback_type"`
	LSEnabled bool `json:"ls_enabled"`

	CommentsEnabled bool `json:"comments_enabled"`

	ExtraEnabled bool `json:"extra_enabled"`

	ExtraField string `json:"extra_field,omitempty"`

	CreationDate string `json:"creation_date,omitempty"`

	//GeoPosition *GeoPosition `json:"geo_position,omitempty"`

	Status string `json:"status,omitempty"`

	PathesToPhoto []AdPhoto `json:"pathes_to_photo,omitempty"`

	Category string `json:"category"`

	CommentsCount int32 `json:"comments_count,omitempty"`
	Hidden        bool  `json:"hidden"`
}

type AdForUsersDetailed struct {
	AdId int64 `json:"ad_id,omitempty"`

	Author *User `json:"author,omitempty"`

	Header string `json:"header"`

	Text string `json:"text"`

	Region string `json:"region"`

	District string `json:"district,omitempty"`

	AdType string `json:"ad_type"`

	//FeedbackType string `json:"feedback_type"`
	LSEnabled bool `json:"ls_enabled"`

	CommentsEnabled bool `json:"comments_enabled"`

	ExtraEnabled bool `json:"extra_enabled"`

	ExtraField string `json:"extra_field,omitempty"`

	CreationDate string `json:"creation_date,omitempty"`

	GeoPosition *GeoPosition `json:"geo_position,omitempty"`

	Status string `json:"status,omitempty"`

	PathesToPhoto []AdPhoto `json:"pathes_to_photo,omitempty"`

	Category string `json:"category"`

	CommentsCount int32 `json:"comments_count,omitempty"`
	ViewsCount    int32 `json:"views_count"`
	Hidden        bool  `json:"hidden"`
	SubscribersNum int32 `json:"subscribers_num"`
}

type AdPhoto struct {
	AdPhotoId int
	PhotoUrl  string
}
