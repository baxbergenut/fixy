package models

import "time"

type Tablet struct {
	ID              string    `json:"id"`
	TruckID         *string   `json:"truck_id,omitempty"`
	TruckUnitNumber *string   `json:"truck_unit_number,omitempty"`
	IMEI            *string   `json:"imei,omitempty"`
	PhoneNumber     *string   `json:"phone_number,omitempty"`
	DeviceMake      *string   `json:"device_make,omitempty"`
	DeviceModel     *string   `json:"device_model,omitempty"`
	ContractType    *string   `json:"contract_type,omitempty"`
	ContractStart   *string   `json:"contract_start,omitempty"`
	ContractEnd     *string   `json:"contract_end,omitempty"`
	Status          *string   `json:"status,omitempty"`
	Notes           *string   `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
