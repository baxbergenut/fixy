package models

import "time"

type Trailer struct {
	ID           string    `json:"id"`
	UnitNumber   string    `json:"unit_number"`
	Vin          *string   `json:"vin,omitempty"`
	PlateNumber  *string   `json:"plate_number,omitempty"`
	Year         *int      `json:"year,omitempty"`
	Make         *string   `json:"make,omitempty"`
	UsageType    *string   `json:"usage_type,omitempty"`
	Location     *string   `json:"location,omitempty"`
	Availability *string   `json:"availability,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
