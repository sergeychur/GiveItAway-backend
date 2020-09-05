package models

type Profile struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Online int `json:"online"`
	City City `json:"city"`
	BDate string `json:"bdate"`
	MobilePhone string `json:"mobile_phone"`
	HomePhone string `json:"home_phone"`
	LastSeen LastSeen `json:"last_seen"`
	CanWritePrivateMessages	int `json:"can_write_private_messages"`
}

type City struct {
	ID int64 `json:"id"`
	Title string `json:"title"`
}

type LastSeen struct {
	Time int `json:"time"`
	Platform int `json:"platform"`
}