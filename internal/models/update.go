package models

type AdUpdate struct {
	Type string	`json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type AdNewStatus struct {
	NewStatus string `json:"new_status"`
}