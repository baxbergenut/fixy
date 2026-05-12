package models

import "time"

type Transponder struct {
	ID                   string    `json:"id"`
	TruckID              *string   `json:"truck_id,omitempty"`
	TruckUnitNumber      *string   `json:"truck_unit_number,omitempty"`
	TransponderNumber    *string   `json:"transponder_number,omitempty"`
	OldTransponderNumber *string   `json:"old_transponder_number,omitempty"`
	MCCompany            *string   `json:"mc_company,omitempty"`
	Status               string    `json:"status"`
	Notes                *string   `json:"notes,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
