package models

import "time"

type Truck struct {
	ID                     string    `json:"id"`
	UnitNumber             string    `json:"unit_number"`
	Vin                    *string   `json:"vin,omitempty"`
	Year                   *int      `json:"year,omitempty"`
	Make                   *string   `json:"make,omitempty"`
	Company                *string   `json:"company,omitempty"`
	Ownership              *string   `json:"ownership,omitempty"`
	PlateNumber            *string   `json:"plate_number,omitempty"`
	PlateState             *string   `json:"plate_state,omitempty"`
	Status                 string    `json:"status"`
	StatusChangedAt        *string   `json:"status_changed_at,omitempty"`
	StatusNote             *string   `json:"status_note,omitempty"`
	SamsaraID              *string   `json:"samsara_id,omitempty"`
	DotInspectionExpiresAt *string   `json:"dot_inspection_expires_at,omitempty"`
	DotInspectionFormURL   *string   `json:"dot_inspection_form_url,omitempty"`
	NextPMOdometer         *int      `json:"next_pm_odometer,omitempty"`
	NextOilChangeOdometer  *int      `json:"next_oil_change_odometer,omitempty"`
	Notes                  *string   `json:"notes,omitempty"`
	Active                 bool      `json:"active"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
