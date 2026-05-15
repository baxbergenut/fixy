package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"fixy-backend/internal/db"
	"fixy-backend/internal/middleware"
	"fixy-backend/internal/models"
)

type TabletsHandler struct {
	database *sql.DB
}

type tabletCreateRequest struct {
	TruckID       *string `json:"truck_id"`
	IMEI          *string `json:"imei"`
	PhoneNumber   *string `json:"phone_number"`
	DeviceMake    *string `json:"device_make"`
	DeviceModel   *string `json:"device_model"`
	ContractType  *string `json:"contract_type"`
	ContractStart *string `json:"contract_start"`
	ContractEnd   *string `json:"contract_end"`
	Status        *string `json:"status"`
	Notes         *string `json:"notes"`
}

type tabletPatchRequest struct {
	TruckID       *string `json:"truck_id"`
	IMEI          *string `json:"imei"`
	PhoneNumber   *string `json:"phone_number"`
	DeviceMake    *string `json:"device_make"`
	DeviceModel   *string `json:"device_model"`
	ContractType  *string `json:"contract_type"`
	ContractStart *string `json:"contract_start"`
	ContractEnd   *string `json:"contract_end"`
	Status        *string `json:"status"`
	Notes         *string `json:"notes"`
}

func NewTabletsHandler(database *sql.DB) http.Handler {
	return &TabletsHandler{database: database}
}

func (handler *TabletsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler.database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/tablets")
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
	case strings.HasPrefix(path, "/"):
		id := strings.TrimPrefix(path, "/")
		if id == "" || strings.Contains(id, "/") {
			notFound(w)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handler.show(w, r, id)
		case http.MethodPatch:
			handler.update(w, r, id)
		default:
			methodNotAllowed(w, http.MethodGet, http.MethodPatch)
		}
	default:
		notFound(w)
	}
}

func (handler *TabletsHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := handler.database.QueryContext(r.Context(), db.ListTabletsQuery)
	if err != nil {
		serverError(w, err)
		return
	}
	defer func() { _ = rows.Close() }()

	items := make([]models.Tablet, 0)
	for rows.Next() {
		item, scanErr := scanTablet(rows)
		if scanErr != nil {
			serverError(w, scanErr)
			return
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (handler *TabletsHandler) show(w http.ResponseWriter, r *http.Request, id string) {
	row := handler.database.QueryRowContext(r.Context(), db.GetTabletQuery, id)
	item, err := scanTablet(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (handler *TabletsHandler) create(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRole(w, r, middleware.RoleAdmin, middleware.RoleFleetManager); !ok {
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request tabletCreateRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	id, err := handler.insertTablet(r, request)
	if err != nil {
		serverError(w, err)
		return
	}

	handler.show(w, r, id)
}

func (handler *TabletsHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := requireRole(w, r, middleware.RoleAdmin, middleware.RoleFleetManager); !ok {
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request tabletPatchRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	updatedID, err := handler.patchTablet(r, id, request)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		serverError(w, err)
		return
	}

	handler.show(w, r, updatedID)
}

func (handler *TabletsHandler) insertTablet(r *http.Request, request tabletCreateRequest) (string, error) {
	contractStart, err := parseDatePtr(request.ContractStart)
	if err != nil {
		return "", err
	}
	contractEnd, err := parseDatePtr(request.ContractEnd)
	if err != nil {
		return "", err
	}

	row := handler.database.QueryRowContext(r.Context(), db.CreateTabletQuery,
		nullString(request.TruckID),
		nullString(request.IMEI),
		nullString(request.PhoneNumber),
		nullString(request.DeviceMake),
		nullString(request.DeviceModel),
		nullString(request.ContractType),
		contractStart,
		contractEnd,
		nullString(request.Status),
		nullString(request.Notes),
	)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (handler *TabletsHandler) patchTablet(r *http.Request, id string, request tabletPatchRequest) (string, error) {
	contractStart, err := parseDatePtr(request.ContractStart)
	if err != nil {
		return "", err
	}
	contractEnd, err := parseDatePtr(request.ContractEnd)
	if err != nil {
		return "", err
	}

	row := handler.database.QueryRowContext(r.Context(), db.UpdateTabletQuery,
		nullString(request.TruckID),
		nullString(request.IMEI),
		nullString(request.PhoneNumber),
		nullString(request.DeviceMake),
		nullString(request.DeviceModel),
		nullString(request.ContractType),
		contractStart,
		contractEnd,
		nullString(request.Status),
		nullString(request.Notes),
		id,
	)

	var updatedID string
	if err := row.Scan(&updatedID); err != nil {
		return "", err
	}

	return updatedID, nil
}

func scanTablet(scanner rowScanner) (models.Tablet, error) {
	var item models.Tablet
	var truckID, truckUnitNumber, imei, phoneNumber, deviceMake, deviceModel, contractType, status, notes sql.NullString
	var contractStart, contractEnd sql.NullTime

	if err := scanner.Scan(
		&item.ID,
		&truckID,
		&truckUnitNumber,
		&imei,
		&phoneNumber,
		&deviceMake,
		&deviceModel,
		&contractType,
		&contractStart,
		&contractEnd,
		&status,
		&notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return models.Tablet{}, err
	}

	item.TruckID = stringPtrFromNull(truckID)
	item.TruckUnitNumber = stringPtrFromNull(truckUnitNumber)
	item.IMEI = stringPtrFromNull(imei)
	item.PhoneNumber = stringPtrFromNull(phoneNumber)
	item.DeviceMake = stringPtrFromNull(deviceMake)
	item.DeviceModel = stringPtrFromNull(deviceModel)
	item.ContractType = stringPtrFromNull(contractType)
	item.ContractStart = dateStringPtrFromNull(contractStart)
	item.ContractEnd = dateStringPtrFromNull(contractEnd)
	item.Status = stringPtrFromNull(status)
	item.Notes = stringPtrFromNull(notes)

	return item, nil
}
