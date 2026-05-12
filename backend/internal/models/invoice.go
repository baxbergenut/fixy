package models

type InvoiceParseResult struct {
	Vendor          *string  `json:"vendor,omitempty"`
	ExpenseDate     *string  `json:"expense_date,omitempty"`
	TruckUnitNumber *string  `json:"truck_unit_number,omitempty"`
	DriverName      *string  `json:"driver_name,omitempty"`
	Amount          *float64 `json:"amount,omitempty"`
	Category        *string  `json:"category,omitempty"`
	Description     *string  `json:"description,omitempty"`
	ReferenceNumber *string  `json:"reference_number,omitempty"`
}
