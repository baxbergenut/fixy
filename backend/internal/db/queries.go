package db

const PingQuery = "SELECT 1"

const TruckColumns = `
	id,
	unit_number,
	vin,
	year,
	make,
	company,
	ownership,
	plate_number,
	plate_state,
	status,
	status_changed_at,
	status_note,
	samsara_id,
	dot_inspection_expires_at,
	dot_inspection_form_url,
	next_pm_odometer,
	next_oil_change_odometer,
	notes,
	active,
	created_at,
	updated_at`

const ListTrucksQuery = `
SELECT
	` + TruckColumns + `
FROM trucks
ORDER BY active DESC, unit_number`

const GetTruckQuery = `
SELECT
	` + TruckColumns + `
FROM trucks
WHERE id = $1`

const CreateTruckQuery = `
INSERT INTO trucks (
	unit_number,
	vin,
	year,
	make,
	company,
	ownership,
	plate_number,
	plate_state,
	status,
	status_changed_at,
	status_note,
	samsara_id,
	dot_inspection_expires_at,
	dot_inspection_form_url,
	next_pm_odometer,
	next_oil_change_odometer,
	notes
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, 'ENROUTE'), $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING
	` + TruckColumns

const UpdateTruckQuery = `
UPDATE trucks
SET
	unit_number = COALESCE($1, unit_number),
	vin = COALESCE($2, vin),
	year = COALESCE($3, year),
	make = COALESCE($4, make),
	company = COALESCE($5, company),
	ownership = COALESCE($6, ownership),
	plate_number = COALESCE($7, plate_number),
	plate_state = COALESCE($8, plate_state),
	status = COALESCE($9, status),
	status_changed_at = COALESCE($10, status_changed_at),
	status_note = COALESCE($11, status_note),
	samsara_id = COALESCE($12, samsara_id),
	dot_inspection_expires_at = COALESCE($13, dot_inspection_expires_at),
	dot_inspection_form_url = COALESCE($14, dot_inspection_form_url),
	next_pm_odometer = COALESCE($15, next_pm_odometer),
	next_oil_change_odometer = COALESCE($16, next_oil_change_odometer),
	notes = COALESCE($17, notes),
	updated_at = NOW()
WHERE id = $18
RETURNING
	` + TruckColumns

const SoftDeleteTruckQuery = `
UPDATE trucks
SET active = FALSE,
		updated_at = NOW()
WHERE id = $1
RETURNING
	` + TruckColumns

const MaintenanceLogColumns = `
	ml.id,
	ml.truck_id,
	t.unit_number AS truck_unit_number,
	ml.trailer_id,
	tr.unit_number AS trailer_unit_number,
	ml.expense_date,
	ml.week_label,
	ml.driver_name,
	ml.amount,
	ml.category,
	ml.payment_type,
	ml.description,
	ml.reference_number,
	ml.who_covers,
	ml.paid_by,
	ml.manager_verified,
	ml.accounting_verified,
	ml.invoice_file_url,
	ml.created_at,
	ml.updated_at`

const ListMaintenanceLogsQuery = `
SELECT
	` + MaintenanceLogColumns + `
FROM maintenance_logs ml
LEFT JOIN trucks t ON t.id = ml.truck_id
LEFT JOIN trailers tr ON tr.id = ml.trailer_id
WHERE ($1::uuid IS NULL OR ml.truck_id = $1::uuid)
ORDER BY ml.expense_date DESC, ml.created_at DESC`

const GetMaintenanceLogQuery = `
SELECT
	` + MaintenanceLogColumns + `
FROM maintenance_logs ml
LEFT JOIN trucks t ON t.id = ml.truck_id
LEFT JOIN trailers tr ON tr.id = ml.trailer_id
WHERE ml.id = $1`

const InsertMaintenanceLogQuery = `
INSERT INTO maintenance_logs (
	truck_id,
	trailer_id,
	expense_date,
	week_label,
	driver_name,
	amount,
	category,
	payment_type,
	description,
	reference_number,
	who_covers,
	paid_by,
	manager_verified,
	accounting_verified,
	invoice_file_url
)
VALUES (
	$1, $2, $3, $4, $5, $6, COALESCE($7::expense_category, 'Other'::expense_category), $8, $9, $10, $11, $12, COALESCE($13, FALSE), COALESCE($14, FALSE), $15
)
RETURNING id`
