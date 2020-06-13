package models

type DonationRequest struct {
	DonationId int32 `json:"donation_id,omitempty"`

	AuthorId int32 `json:"author_id,omitempty"`

	SumNeeded int32 `json:"sum_needed,omitempty"`

	ExpireDate string `json:"expire_date,omitempty"`

	Status string `json:"status,omitempty"`

	Description string `json:"description,omitempty"`
}
