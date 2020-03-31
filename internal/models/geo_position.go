package models

type GeoPosition struct {
	Available bool    `json:"available,omitempty"`
	Latitude  string `json:"lat,omitempty"`
	Longitude string `json:"long,omitempty"`
}
