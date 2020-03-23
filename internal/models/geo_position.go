package models

type GeoPosition struct {
	Available bool    `json:"available,omitempty"`
	Latitude  float64 `json:"lat,omitempty"`
	Longitude float64 `json:"long,omitempty"`
}
