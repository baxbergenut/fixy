package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fixy-backend/internal/db"
	"fixy-backend/internal/models"
)

type TrucksHandler struct {
	database *sql.DB
}

type truckCreateRequest struct {
	UnitNumber             string  `json:"unit_number"`
	Vin                    *string `json:"vin"`
	Year                   *int    `json:"year"`
	Make                   *string `json:"make"`
	Company                *string `json:"company"`
	Ownership              *string `json:"ownership"`
	PlateNumber            *string `json:"plate_number"`
	PlateState             *string `json:"plate_state"`
	Status                 *string `json:"status"`
	StatusChangedAt        *string `json:"status_changed_at"`
	StatusNote             *string `json:"status_note"`
	SamsaraID              *string `json:"samsara_id"`
	DotInspectionExpiresAt *string `json:"dot_inspection_expires_at"`
	DotInspectionFormURL   *string `json:"dot_inspection_form_url"`
	NextPMOdometer         *int    `json:"next_pm_odometer"`
	NextOilChangeOdometer  *int    `json:"next_oil_change_odometer"`
	Notes                  *string `json:"notes"`
}

type truckPatchRequest struct {
	UnitNumber             *string `json:"unit_number"`
	Vin                    *string `json:"vin"`
	Year                   *int    `json:"year"`
	Make                   *string `json:"make"`
	Company                *string `json:"company"`
	Ownership              *string `json:"ownership"`
	PlateNumber            *string `json:"plate_number"`
	PlateState             *string `json:"plate_state"`
	Status                 *string `json:"status"`
	StatusChangedAt        *string `json:"status_changed_at"`
	StatusNote             *string `json:"status_note"`
	SamsaraID              *string `json:"samsara_id"`
	DotInspectionExpiresAt *string `json:"dot_inspection_expires_at"`
	DotInspectionFormURL   *string `json:"dot_inspection_form_url"`
	NextPMOdometer         *int    `json:"next_pm_odometer"`
	NextOilChangeOdometer  *int    `json:"next_oil_change_odometer"`
	Notes                  *string `json:"notes"`
}

func NewTrucksHandler(database *sql.DB) http.Handler {
	return &TrucksHandler{database: database}
}

func (handler *TrucksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler.database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/trucks")
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
		case http.MethodDelete:
			handler.softDelete(w, r, id)
		default:
			methodNotAllowed(w, http.MethodGet, http.MethodPatch, http.MethodDelete)
		}
	default:
		notFound(w)
	}
}

func (handler *TrucksHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := handler.database.QueryContext(r.Context(), db.ListTrucksQuery)
	if err != nil {
		serverError(w, err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()

	trucks := make([]models.Truck, 0)
	for rows.Next() {
		truck, scanErr := scanTruck(rows)
		if scanErr != nil {
			serverError(w, scanErr)
			return
		}
		trucks = append(trucks, truck)
	}

	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, trucks)
}

func (handler *TrucksHandler) show(w http.ResponseWriter, r *http.Request, id string) {
	row := handler.database.QueryRowContext(r.Context(), db.GetTruckQuery, id)
	truck, err := scanTruck(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, truck)
}

func (handler *TrucksHandler) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request truckCreateRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	if strings.TrimSpace(request.UnitNumber) == "" {
		badRequest(w, errors.New("unit_number is required"))
		return
	}

	createdTruck, err := handler.insertTruck(r, request)
	if err != nil {
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createdTruck)
}

func (handler *TrucksHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request truckPatchRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	updatedTruck, err := handler.patchTruck(r, id, request)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updatedTruck)
}

func (handler *TrucksHandler) softDelete(w http.ResponseWriter, r *http.Request, id string) {
	row := handler.database.QueryRowContext(r.Context(), db.SoftDeleteTruckQuery, id)
	truck, err := scanTruck(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, truck)
}

func (handler *TrucksHandler) insertTruck(r *http.Request, request truckCreateRequest) (models.Truck, error) {
	params, err := buildTruckCreateParams(request)
	if err != nil {
		return models.Truck{}, err
	}

	row := handler.database.QueryRowContext(r.Context(), db.CreateTruckQuery,
		params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7], params[8], params[9], params[10], params[11], params[12], params[13], params[14], params[15], params[16],
	)

	return scanTruck(row)
}

func (handler *TrucksHandler) patchTruck(r *http.Request, id string, request truckPatchRequest) (models.Truck, error) {
	params, err := buildTruckPatchParams(request, id)
	if err != nil {
		return models.Truck{}, err
	}

	row := handler.database.QueryRowContext(r.Context(), db.UpdateTruckQuery,
		params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7], params[8], params[9], params[10], params[11], params[12], params[13], params[14], params[15], params[16], params[17],
	)

	return scanTruck(row)
}

func buildTruckCreateParams(request truckCreateRequest) ([]any, error) {
	status := "ENROUTE"
	if request.Status != nil && strings.TrimSpace(*request.Status) != "" {
		status = strings.TrimSpace(*request.Status)
	}

	statusChangedAt, err := parseDatePtr(request.StatusChangedAt)
	if err != nil {
		return nil, err
	}
	dotInspectionExpiresAt, err := parseDatePtr(request.DotInspectionExpiresAt)
	if err != nil {
		return nil, err
	}

	params := []any{
		strings.TrimSpace(request.UnitNumber),
		nullString(request.Vin),
		nullInt(request.Year),
		nullString(request.Make),
		nullString(request.Company),
		nullString(request.Ownership),
		nullString(request.PlateNumber),
		nullString(request.PlateState),
		status,
		statusChangedAt,
		nullString(request.StatusNote),
		nullString(request.SamsaraID),
		dotInspectionExpiresAt,
		nullString(request.DotInspectionFormURL),
		nullInt(request.NextPMOdometer),
		nullInt(request.NextOilChangeOdometer),
		nullString(request.Notes),
	}

	return params, nil
}

func buildTruckPatchParams(request truckPatchRequest, id string) ([]any, error) {
	statusChangedAt, err := parseDatePtr(request.StatusChangedAt)
	if err != nil {
		return nil, err
	}
	dotInspectionExpiresAt, err := parseDatePtr(request.DotInspectionExpiresAt)
	if err != nil {
		return nil, err
	}

	params := []any{
		nullString(request.UnitNumber),
		nullString(request.Vin),
		nullInt(request.Year),
		nullString(request.Make),
		nullString(request.Company),
		nullString(request.Ownership),
		nullString(request.PlateNumber),
		nullString(request.PlateState),
		nullString(request.Status),
		statusChangedAt,
		nullString(request.StatusNote),
		nullString(request.SamsaraID),
		dotInspectionExpiresAt,
		nullString(request.DotInspectionFormURL),
		nullInt(request.NextPMOdometer),
		nullInt(request.NextOilChangeOdometer),
		nullString(request.Notes),
		id,
	}

	return params, nil
}

func scanTruck(scanner rowScanner) (models.Truck, error) {
	var truck models.Truck
	var vin, makeValue, company, ownership, plateNumber, plateState, status, statusNote, samsaraID, dotFormURL, notes sql.NullString
	var year sql.NullInt32
	var nextPM, nextOil sql.NullInt32
	var statusChangedAt, dotInspectionExpiresAt sql.NullTime

	if err := scanner.Scan(
		&truck.ID,
		&truck.UnitNumber,
		&vin,
		&year,
		&makeValue,
		&company,
		&ownership,
		&plateNumber,
		&plateState,
		&status,
		&statusChangedAt,
		&statusNote,
		&samsaraID,
		&dotInspectionExpiresAt,
		&dotFormURL,
		&nextPM,
		&nextOil,
		&notes,
		&truck.Active,
		&truck.CreatedAt,
		&truck.UpdatedAt,
	); err != nil {
		return models.Truck{}, err
	}

	truck.Vin = stringPtrFromNull(vin)
	truck.Year = intPtrFromNull(year)
	truck.Make = stringPtrFromNull(makeValue)
	truck.Company = stringPtrFromNull(company)
	truck.Ownership = stringPtrFromNull(ownership)
	truck.PlateNumber = stringPtrFromNull(plateNumber)
	truck.PlateState = stringPtrFromNull(plateState)
	truck.Status = status.String
	truck.StatusChangedAt = dateStringPtrFromNull(statusChangedAt)
	truck.StatusNote = stringPtrFromNull(statusNote)
	truck.SamsaraID = stringPtrFromNull(samsaraID)
	truck.DotInspectionExpiresAt = dateStringPtrFromNull(dotInspectionExpiresAt)
	truck.DotInspectionFormURL = stringPtrFromNull(dotFormURL)
	truck.NextPMOdometer = intPtrFromNull(nextPM)
	truck.NextOilChangeOdometer = intPtrFromNull(nextOil)
	truck.Notes = stringPtrFromNull(notes)

	return truck, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}

func parseDatePtr(value *string) (any, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", trimmed, err)
	}

	return parsed, nil
}

func nullString(value *string) any {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}

func nullInt(value *int) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullBool(value *bool) any {
	if value == nil {
		return nil
	}

	return *value
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	text := value.String
	return &text
}

func intPtrFromNull(value sql.NullInt32) *int {
	if !value.Valid {
		return nil
	}

	result := int(value.Int32)
	return &result
}

func dateStringPtrFromNull(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}

	text := value.Time.UTC().Format("2006-01-02")
	return &text
}

func methodNotAllowed(w http.ResponseWriter, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func badRequest(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
}

func serverError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
