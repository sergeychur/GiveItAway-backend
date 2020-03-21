package models

type User struct {

	VkId int64 `json:"vk_id,omitempty"`

	Carma int32 `json:"carma,omitempty"`

	Name string `json:"name,omitempty"`

	Surname string `json:"surname,omitempty"`

	PhotoUrl string `json:"photo_url,omitempty"`
}