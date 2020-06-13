package models

type Donation struct {
	UserId int32 `json:"user_id,omitempty"`

	DonatId int32 `json:"donat_id,omitempty"`

	Region string `json:"region,omitempty"`

	District string `json:"district,omitempty"`

	Sum int32 `json:"sum,omitempty"`

	Category string `json:"category,omitempty"`
}
