package models

type User struct {
	VkId int64 `json:"vk_id,omitempty"`

	Carma int32 `json:"carma,omitempty"`

	Name string `json:"name,omitempty"`

	Surname string `json:"surname,omitempty"`

	PhotoUrl string `json:"photo_url,omitempty"`
}

type UserForProfile struct {
	VkId int64 `json:"vk_id,omitempty"`

	Carma int32 `json:"carma,omitempty"`

	Name string `json:"name,omitempty"`

	Surname string `json:"surname,omitempty"`

	PhotoUrl string `json:"photo_url,omitempty"`

	RegistrationDate string `json:"registration_date"`

	FrozenCarma int `json:"frozen_carma"`

	TotalEarnedCarma int `json:"total_earned_carma"`

	TotalSpentCarma int `json:"total_spent_carma"`

	TotalGivenAds int `json:"total_given_ads"`

	TotalReceivedAds int `json:"total_received_ads"`

	TotalAbortedAds int `json:"total_aborted_ads"`
}

type CanSend struct {
	CanSend bool `json:"can_send"`
}
