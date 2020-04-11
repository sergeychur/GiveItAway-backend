package models

type Notification struct {
	NotificationType string `json:"notification_type"`
	WhomId int /*`json:"whom_id,omitempty"`*/
	CreationDateTime string `json:"creation_date_time"`
	Payload interface{} `json:"payload"`
	IsRead bool `json:"is_read"`
}

type AdForNotification struct {
	AdId int64 `json:"ad_id"`
	Status string `json:"status"`
	Header string `json:"header"`
	PathesToPhoto []AdPhoto `json:"pathes_to_photo,omitempty"`
}

type AuthorClosedAd struct {
	Author User `json:"author"`
	Ad AdForNotification `json:"ad"`
	DealId int `json:"deal_id"`
}

type AdStatusChanged struct {
	Ad AdForNotification `json:"ad"`
}	// for change status, delete, user fulfilled

type UserSubscribed struct {
	Ad AdForNotification `json:"ad"`
	Author User `json:"author"`
}