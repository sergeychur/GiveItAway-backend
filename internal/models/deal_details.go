package models

type DealDetails struct {

	DealId int32 `json:"deal_id,omitempty"`

	AdId int32 `json:"ad_id,omitempty"`

	AuthorId int32 `json:"author_id,omitempty"`

	SubscriberId string `json:"subscriber_id,omitempty"`

	Status string `json:"status,omitempty"`
}