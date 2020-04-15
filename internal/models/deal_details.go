package models

type DealDetails struct {
	DealId int32 `json:"deal_id,omitempty"`

	AdId int32 `json:"ad_id,omitempty"`

	SubscriberId int32 `json:"subscriber_id,omitempty"`

	Status string `json:"status,omitempty"`
}

type Bid struct {
	Bid int `json:"bid"`
}