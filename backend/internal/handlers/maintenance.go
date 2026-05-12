package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"fixy-backend/internal/db"
	"fixy-backend/internal/models"
	"fixy-backend/internal/services"
)

type MaintenanceHandler struct {
	database *sql.DB
}

func NewMaintenanceHandler(database *sql.DB) http.Handler {
	return &MaintenanceHandler{database: database}
}

func (handler *MaintenanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler.database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/maintenance")
	switch {
	case path == "" || path == "/":
		switch r.Method {
		case http.MethodGet:
			handler.list(w, r)
		case http.MethodPost:
			handler.create(w, r)
		default:
			methodNotAllowed(w, http.MethodGet, http.MethodPost)
		}
	default:
		notFound(w)
	}
}

func (handler *MaintenanceHandler) list(w http.ResponseWriter, r *http.Request) {
	var truckID any
	if queryTruckID := strings.TrimSpace(r.URL.Query().Get("truck_id")); queryTruckID != "" {
		truckID = queryTruckID
	}

	rows, err := handler.database.QueryContext(r.Context(), db.ListMaintenanceLogsQuery, truckID)
	if err != nil {
		serverError(w, err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()

	logs := make([]models.MaintenanceLog, 0)
	for rows.Next() {
		logEntry, scanErr := scanMaintenanceLog(rows)
		if scanErr != nil {
			serverError(w, scanErr)
			return
		}
		logs = append(logs, logEntry)
	}

	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

func (handler *MaintenanceHandler) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request maintenanceCreateRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	if strings.TrimSpace(request.ExpenseDate) == "" {
		badRequest(w, errors.New("expense_date is required"))
		return
	}

	if request.Amount == nil {
		badRequest(w, errors.New("amount is required"))
		return
	}

	truckID := stringValue(request.TruckID)
	trailerID := stringValue(request.TrailerID)
	if truckID == "" && trailerID == "" {
		badRequest(w, errors.New("truck_id or trailer_id is required"))
		return
	}

	expenseDate, err := parseDatePtr(&request.ExpenseDate)
	if err != nil {
		badRequest(w, err)
		return
	}

	category := services.NormalizeMaintenanceCategory(request.Category)
	managerVerified := false
	if request.ManagerVerified != nil {
		managerVerified = *request.ManagerVerified
	}
	accountingVerified := false
	if request.AccountingVerified != nil {
		accountingVerified = *request.AccountingVerified
	}

	row := handler.database.QueryRowContext(r.Context(), db.InsertMaintenanceLogQuery,
		nullString(request.TruckID),
		nullString(request.TrailerID),
		expenseDate,
		nullString(request.WeekLabel),
		nullString(request.DriverName),
		*request.Amount,
		category,
		nullString(request.PaymentType),
		nullString(request.Description),
		nullString(request.ReferenceNumber),
		nullString(request.WhoCovers),
		nullString(request.PaidBy),
		managerVerified,
		accountingVerified,
		nullString(request.InvoiceFileURL),
	)

	var createdID string
	if err := row.Scan(&createdID); err != nil {
		serverError(w, err)
		return
	}

	createdRow := handler.database.QueryRowContext(r.Context(), db.GetMaintenanceLogQuery, createdID)
	createdLog, err := scanMaintenanceLog(createdRow)
	if err != nil {
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createdLog)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}

func scanMaintenanceLog(scanner rowScanner) (models.MaintenanceLog, error) {
	var logEntry models.MaintenanceLog
	var truckID, truckUnitNumber, trailerID, trailerUnitNumber, weekLabel, driverName, paymentType, description, referenceNumber, whoCovers, paidBy, invoiceFileURL sql.NullString
	var amount sql.NullFloat64
	var expenseDate sql.NullTime
	var createdAt, updatedAt time.Time

	if err := scanner.Scan(
		&logEntry.ID,
		&truckID,
		&truckUnitNumber,
		&trailerID,
		&trailerUnitNumber,
		&expenseDate,
		&weekLabel,
		&driverName,
		&amount,
		&logEntry.Category,
		&paymentType,
		&description,
		&referenceNumber,
		&whoCovers,
		&paidBy,
		&logEntry.ManagerVerified,
		&logEntry.AccountingVerified,
		&invoiceFileURL,
		&createdAt,
		&updatedAt,
	); err != nil {
		return models.MaintenanceLog{}, err
	}

	if truckID.Valid {
		value := truckID.String
		logEntry.TruckID = &value
	}
	if truckUnitNumber.Valid {
		value := truckUnitNumber.String
		logEntry.TruckUnitNumber = &value
	}
	if trailerID.Valid {
		value := trailerID.String
		logEntry.TrailerID = &value
	}
	if trailerUnitNumber.Valid {
		value := trailerUnitNumber.String
		logEntry.TrailerUnitNumber = &value
	}
	if weekLabel.Valid {
		value := weekLabel.String
		logEntry.WeekLabel = &value
	}
	if driverName.Valid {
		value := driverName.String
		logEntry.DriverName = &value
	}
	if amount.Valid {
		logEntry.Amount = amount.Float64
	}
	if paymentType.Valid {
		value := paymentType.String
		logEntry.PaymentType = &value
	}
	if description.Valid {
		value := description.String
		logEntry.Description = &value
	}
	if referenceNumber.Valid {
		value := referenceNumber.String
		logEntry.ReferenceNumber = &value
	}
	if whoCovers.Valid {
		value := whoCovers.String
		logEntry.WhoCovers = &value
	}
	if paidBy.Valid {
		value := paidBy.String
		logEntry.PaidBy = &value
	}
	if invoiceFileURL.Valid {
		value := invoiceFileURL.String
		logEntry.InvoiceFileURL = &value
	}
	if expenseDate.Valid {
		logEntry.ExpenseDate = expenseDate.Time.UTC().Format("2006-01-02")
	}
	logEntry.CreatedAt = createdAt
	logEntry.UpdatedAt = updatedAt

	return logEntry, nil
}

type maintenanceCreateRequest struct {
	TruckID            *string  `json:"truck_id"`
	TrailerID          *string  `json:"trailer_id"`
	ExpenseDate        string   `json:"expense_date"`
	WeekLabel          *string  `json:"week_label"`
	DriverName         *string  `json:"driver_name"`
	Amount             *float64 `json:"amount"`
	Category           string   `json:"category"`
	PaymentType        *string  `json:"payment_type"`
	Description        *string  `json:"description"`
	ReferenceNumber    *string  `json:"reference_number"`
	WhoCovers          *string  `json:"who_covers"`
	PaidBy             *string  `json:"paid_by"`
	ManagerVerified    *bool    `json:"manager_verified"`
	AccountingVerified *bool    `json:"accounting_verified"`
	InvoiceFileURL     *string  `json:"invoice_file_url"`
}
