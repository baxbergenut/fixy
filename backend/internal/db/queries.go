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

const TrailerColumns = `
	id,
	unit_number,
	vin,
	plate_number,
	year,
	make,
	usage_type,
	location,
	availability,
	notes,
	created_at,
	updated_at`

const TransponderColumns = `
	transponders.id,
	transponders.truck_id,
	t.unit_number AS truck_unit_number,
	transponders.transponder_number,
	transponders.old_transponder_number,
	transponders.mc_company,
	transponders.status,
	transponders.notes,
	transponders.created_at,
	transponders.updated_at`

const TabletColumns = `
	tablets.id,
	tablets.truck_id,
	t.unit_number AS truck_unit_number,
	tablets.imei,
	tablets.phone_number,
	tablets.device_make,
	tablets.device_model,
	tablets.contract_type,
	tablets.contract_start,
	tablets.contract_end,
	tablets.status,
	tablets.notes,
	tablets.created_at,
	tablets.updated_at`

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

const GetTruckIDByUnitNumberQuery = `
SELECT id
FROM trucks
WHERE unit_number = $1
	OR (
		unit_number ~ '^[0-9]+$'
		AND $1 ~ '^[0-9]+$'
		AND unit_number::int = $1::int
	)
ORDER BY active DESC, unit_number
LIMIT 1`

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

const ListTrailersQuery = `
SELECT
	` + TrailerColumns + `
FROM trailers
ORDER BY unit_number`

const GetTrailerQuery = `
SELECT
	` + TrailerColumns + `
FROM trailers
WHERE id = $1`

const CreateTrailerQuery = `
INSERT INTO trailers (
	unit_number,
	vin,
	plate_number,
	year,
	make,
	usage_type,
	location,
	availability,
	notes
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id`

const UpdateTrailerQuery = `
UPDATE trailers
SET
	unit_number = COALESCE($1, unit_number),
	vin = COALESCE($2, vin),
	plate_number = COALESCE($3, plate_number),
	year = COALESCE($4, year),
	make = COALESCE($5, make),
	usage_type = COALESCE($6, usage_type),
	location = COALESCE($7, location),
	availability = COALESCE($8, availability),
	notes = COALESCE($9, notes),
	updated_at = NOW()
WHERE id = $10
RETURNING id`

const ListTranspondersQuery = `
SELECT
	` + TransponderColumns + `
FROM transponders
LEFT JOIN trucks t ON t.id = transponders.truck_id
ORDER BY transponders.created_at DESC`

const GetTransponderQuery = `
SELECT
	` + TransponderColumns + `
FROM transponders
LEFT JOIN trucks t ON t.id = transponders.truck_id
WHERE transponders.id = $1`

const CreateTransponderQuery = `
INSERT INTO transponders (
	truck_id,
	transponder_number,
	old_transponder_number,
	mc_company,
	status,
	notes
)
VALUES (
	$1, $2, $3, $4, COALESCE($5, 'Active'::transponder_status), $6
)
RETURNING id`

const UpdateTransponderQuery = `
UPDATE transponders
SET
	truck_id = COALESCE($1, truck_id),
	transponder_number = COALESCE($2, transponder_number),
	old_transponder_number = COALESCE($3, old_transponder_number),
	mc_company = COALESCE($4, mc_company),
	status = COALESCE($5, status),
	notes = COALESCE($6, notes),
	updated_at = NOW()
WHERE id = $7
RETURNING id`

const ListTabletsQuery = `
SELECT
	` + TabletColumns + `
FROM tablets
LEFT JOIN trucks t ON t.id = tablets.truck_id
ORDER BY tablets.created_at DESC`

const GetTabletQuery = `
SELECT
	` + TabletColumns + `
FROM tablets
LEFT JOIN trucks t ON t.id = tablets.truck_id
WHERE tablets.id = $1`

const CreateTabletQuery = `
INSERT INTO tablets (
	truck_id,
	imei,
	phone_number,
	device_make,
	device_model,
	contract_type,
	contract_start,
	contract_end,
	status,
	notes
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id`

const UpdateTabletQuery = `
UPDATE tablets
SET
	truck_id = COALESCE($1, truck_id),
	imei = COALESCE($2, imei),
	phone_number = COALESCE($3, phone_number),
	device_make = COALESCE($4, device_make),
	device_model = COALESCE($5, device_model),
	contract_type = COALESCE($6, contract_type),
	contract_start = COALESCE($7, contract_start),
	contract_end = COALESCE($8, contract_end),
	status = COALESCE($9, status),
	notes = COALESCE($10, notes),
	updated_at = NOW()
WHERE id = $11
RETURNING id`

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
	ml.telegram_message,
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
	telegram_message,
	manager_verified,
	accounting_verified,
	invoice_file_url
)
VALUES (
	$1, $2, $3, $4, $5, $6, COALESCE($7::expense_category, 'Other'::expense_category), $8, $9, $10, $11, $12, $13, COALESCE($14, FALSE), COALESCE($15, FALSE), $16
)
RETURNING id`

const UpdateMaintenanceLogQuery = `
UPDATE maintenance_logs
SET
	truck_id = COALESCE($1, truck_id),
	trailer_id = COALESCE($2, trailer_id),
	expense_date = COALESCE($3, expense_date),
	week_label = COALESCE($4, week_label),
	driver_name = COALESCE($5, driver_name),
	amount = COALESCE($6, amount),
	category = COALESCE($7::expense_category, category),
	payment_type = COALESCE($8, payment_type),
	description = COALESCE($9, description),
	reference_number = COALESCE($10, reference_number),
	who_covers = COALESCE($11, who_covers),
	paid_by = COALESCE($12, paid_by),
	manager_verified = COALESCE($13, manager_verified),
	accounting_verified = COALESCE($14, accounting_verified),
	invoice_file_url = COALESCE($15, invoice_file_url),
	updated_at = NOW()
WHERE id = $16
RETURNING id`
