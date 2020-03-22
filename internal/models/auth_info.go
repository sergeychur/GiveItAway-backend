package models

type AuthInfo struct {
	Url      string `json:"url,omitempty"`
	Name     string `json:"name,omitempty"`
	Surname  string `json:"surname,omitempty"`
	PhotoURL string `json:"photo_url,omitempty"`
}
