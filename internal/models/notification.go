package models

type Notification struct {
	NotificationId int	`json:"notification_id"` 
	NotificationType string      `json:"notification_type"`
	WhomId           int         `json:"-"`/*`json:"whom_id,omitempty"`*/
	CreationDateTime string      `json:"creation_date_time"`
	Payload          interface{} `json:"payload"`
	IsRead           bool        `json:"is_read"`
	AdId int64 `json:"-"`
}

type AdForNotification struct {
	AdId          int64     `json:"ad_id"`
	Status        string    `json:"status"`
	Header        string    `json:"header"`
	PathesToPhoto []AdPhoto `json:"pathes_to_photo,omitempty"`
}

type AuthorClosedAd struct {
	Author User              `json:"author"`
	Ad     AdForNotification `json:"ad"`
	DealId int               `json:"deal_id"`
}

type AdStatusChanged struct {
	Ad AdForNotification `json:"ad"`
} // for change status, delete, user fulfilled

type UserSubscribed struct {
	Ad     AdForNotification `json:"ad"`
	Author User              `json:"author"`
}

type CancelInfo struct {
	WhomId     int
	CancelType string
	AdId       int
}

type SubscriberCancelled struct {
	Ad     AdForNotification `json:"ad"`
	Author User              `json:"author"`
}

type AuthorCancelled struct {
	Ad AdForNotification `json:"ad"`
}

type Unsubscribed struct {
	UserId int `json:"user_id"`
}

type NotesNumber struct {
	Number int `json:"number"`
}

type NewComment struct {
	Ad AdForNotification `json:"ad"`
	Comment CommentForUser `json:"comment"`
}

type CommentId struct {
	CommentId int `json:"comment_id"`
}