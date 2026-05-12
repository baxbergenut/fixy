export type TruckStatus = "ENROUTE" | "SHOP" | "STOP" | "UNAVAILABLE";

export type Truck = {
    id: string;
    unit_number: string;
    vin: string | null;
    year: number | null;
    make: string | null;
    company: string | null;
    ownership: string | null;
    plate_number: string | null;
    plate_state: string | null;
    status: TruckStatus;
    status_changed_at: string | null;
    status_note: string | null;
    samsara_id: string | null;
    dot_inspection_expires_at: string | null;
    dot_inspection_form_url: string | null;
    next_pm_odometer: number | null;
    next_oil_change_odometer: number | null;
    notes: string | null;
    active: boolean;
    created_at: string;
    updated_at: string;
};

export type MaintenanceLog = {
    id: string;
    truck_id: string | null;
    truck_unit_number: string | null;
    trailer_id: string | null;
    trailer_unit_number: string | null;
    expense_date: string;
    week_label: string | null;
    driver_name: string | null;
    amount: number;
    category: string;
    payment_type: string | null;
    description: string | null;
    reference_number: string | null;
    who_covers: string | null;
    paid_by: string | null;
    manager_verified: boolean;
    accounting_verified: boolean;
    invoice_file_url: string | null;
    created_at: string;
    updated_at: string;
};

export type InvoiceParseResult = {
    vendor: string | null;
    expense_date: string | null;
    truck_unit_number: string | null;
    driver_name: string | null;
    amount: number | null;
    category: string | null;
    description: string | null;
    reference_number: string | null;
};

export type MaintenanceCreateRequest = {
    truck_id: string;
    trailer_id?: string | null;
    expense_date: string;
    week_label?: string | null;
    driver_name?: string | null;
    amount: number;
    category: string;
    payment_type?: string | null;
    description?: string | null;
    reference_number?: string | null;
    who_covers?: string | null;
    paid_by?: string | null;
    manager_verified?: boolean;
    accounting_verified?: boolean;
    invoice_file_url?: string | null;
};
