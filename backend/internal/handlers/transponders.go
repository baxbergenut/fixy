package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"fixy-backend/internal/db"
	"fixy-backend/internal/models"
)

type TranspondersHandler struct {
	database *sql.DB
}

type transponderCreateRequest struct {
	TruckID              *string `json:"truck_id"`
	TransponderNumber    *string `json:"transponder_number"`
	OldTransponderNumber *string `json:"old_transponder_number"`
	MCCompany            *string `json:"mc_company"`
	Status               *string `json:"status"`
	Notes                *string `json:"notes"`
}

type transponderPatchRequest struct {
	TruckID              *string `json:"truck_id"`
	TransponderNumber    *string `json:"transponder_number"`
	OldTransponderNumber *string `json:"old_transponder_number"`
	MCCompany            *string `json:"mc_company"`
	Status               *string `json:"status"`
	Notes                *string `json:"notes"`
}

func NewTranspondersHandler(database *sql.DB) http.Handler {
	return &TranspondersHandler{database: database}
}

func (handler *TranspondersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler.database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/transponders")
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

func (handler *TranspondersHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := handler.database.QueryContext(r.Context(), db.ListTranspondersQuery)
	if err != nil {
		serverError(w, err)
		return
	}
	defer func() { _ = rows.Close() }()

	items := make([]models.Transponder, 0)
	for rows.Next() {
		item, scanErr := scanTransponder(rows)
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

func (handler *TranspondersHandler) show(w http.ResponseWriter, r *http.Request, id string) {
	row := handler.database.QueryRowContext(r.Context(), db.GetTransponderQuery, id)
	item, err := scanTransponder(row)
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

func (handler *TranspondersHandler) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request transponderCreateRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	if request.TransponderNumber == nil || strings.TrimSpace(*request.TransponderNumber) == "" {
		badRequest(w, errors.New("transponder_number is required"))
		return
	}

	id, err := handler.insertTransponder(r, request)
	if err != nil {
		serverError(w, err)
		return
	}

	handler.show(w, r, id)
}

func (handler *TranspondersHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request transponderPatchRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	updatedID, err := handler.patchTransponder(r, id, request)
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

func (handler *TranspondersHandler) insertTransponder(r *http.Request, request transponderCreateRequest) (string, error) {
	row := handler.database.QueryRowContext(r.Context(), db.CreateTransponderQuery,
		nullString(request.TruckID),
		nullString(request.TransponderNumber),
		nullString(request.OldTransponderNumber),
		nullString(request.MCCompany),
		nullString(request.Status),
		nullString(request.Notes),
	)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (handler *TranspondersHandler) patchTransponder(r *http.Request, id string, request transponderPatchRequest) (string, error) {
	row := handler.database.QueryRowContext(r.Context(), db.UpdateTransponderQuery,
		nullString(request.TruckID),
		nullString(request.TransponderNumber),
		nullString(request.OldTransponderNumber),
		nullString(request.MCCompany),
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

func scanTransponder(scanner rowScanner) (models.Transponder, error) {
	var item models.Transponder
	var truckID, truckUnitNumber, transponderNumber, oldTransponderNumber, mcCompany, notes sql.NullString
	var status sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&truckID,
		&truckUnitNumber,
		&transponderNumber,
		&oldTransponderNumber,
		&mcCompany,
		&status,
		&notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return models.Transponder{}, err
	}

	item.TruckID = stringPtrFromNull(truckID)
	item.TruckUnitNumber = stringPtrFromNull(truckUnitNumber)
	item.TransponderNumber = stringPtrFromNull(transponderNumber)
	item.OldTransponderNumber = stringPtrFromNull(oldTransponderNumber)
	item.MCCompany = stringPtrFromNull(mcCompany)
	item.Status = status.String
	item.Notes = stringPtrFromNull(notes)

	return item, nil
}
