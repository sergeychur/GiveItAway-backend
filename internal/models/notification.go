package models

type Notification struct {
	NotificationType string `json:"notification_type"`
	CreationDateTime string `json:"creation_date_time"`
	Payload interface{} `json:"payload"`
	IsRead bool `json:"is_read"`
}

type AuthorClosedAd struct {
	Author User `json:"author"`
	AdTitle string `json:"ad_title"`
	DealId int `json:"deal_id"`
}
