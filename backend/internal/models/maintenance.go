package models

import "time"

type MaintenanceLog struct {
	ID                 string    `json:"id"`
	TruckID            *string   `json:"truck_id,omitempty"`
	TruckUnitNumber    *string   `json:"truck_unit_number,omitempty"`
	TrailerID          *string   `json:"trailer_id,omitempty"`
	TrailerUnitNumber  *string   `json:"trailer_unit_number,omitempty"`
	ExpenseDate        string    `json:"expense_date"`
	WeekLabel          *string   `json:"week_label,omitempty"`
	DriverName         *string   `json:"driver_name,omitempty"`
	Amount             float64   `json:"amount"`
	Category           string    `json:"category"`
	PaymentType        *string   `json:"payment_type,omitempty"`
	Description        *string   `json:"description,omitempty"`
	ReferenceNumber    *string   `json:"reference_number,omitempty"`
	WhoCovers          *string   `json:"who_covers,omitempty"`
	PaidBy             *string   `json:"paid_by,omitempty"`
	ManagerVerified    bool      `json:"manager_verified"`
	AccountingVerified bool      `json:"accounting_verified"`
	InvoiceFileURL     *string   `json:"invoice_file_url,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
