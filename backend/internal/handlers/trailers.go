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

type TrailersHandler struct {
	database *sql.DB
}

type trailerCreateRequest struct {
	UnitNumber   string  `json:"unit_number"`
	Vin          *string `json:"vin"`
	PlateNumber  *string `json:"plate_number"`
	Year         *int    `json:"year"`
	Make         *string `json:"make"`
	UsageType    *string `json:"usage_type"`
	Location     *string `json:"location"`
	Availability *string `json:"availability"`
	Notes        *string `json:"notes"`
}

type trailerPatchRequest struct {
	UnitNumber   *string `json:"unit_number"`
	Vin          *string `json:"vin"`
	PlateNumber  *string `json:"plate_number"`
	Year         *int    `json:"year"`
	Make         *string `json:"make"`
	UsageType    *string `json:"usage_type"`
	Location     *string `json:"location"`
	Availability *string `json:"availability"`
	Notes        *string `json:"notes"`
}

func NewTrailersHandler(database *sql.DB) http.Handler {
	return &TrailersHandler{database: database}
}

func (handler *TrailersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler.database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/trailers")
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

func (handler *TrailersHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := handler.database.QueryContext(r.Context(), db.ListTrailersQuery)
	if err != nil {
		serverError(w, err)
		return
	}
	defer func() { _ = rows.Close() }()

	trailers := make([]models.Trailer, 0)
	for rows.Next() {
		trailer, scanErr := scanTrailer(rows)
		if scanErr != nil {
			serverError(w, scanErr)
			return
		}
		trailers = append(trailers, trailer)
	}

	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, trailers)
}

func (handler *TrailersHandler) show(w http.ResponseWriter, r *http.Request, id string) {
	row := handler.database.QueryRowContext(r.Context(), db.GetTrailerQuery, id)
	trailer, err := scanTrailer(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, trailer)
}

func (handler *TrailersHandler) create(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRole(w, r, middleware.RoleAdmin, middleware.RoleFleetManager); !ok {
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request trailerCreateRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	if strings.TrimSpace(request.UnitNumber) == "" {
		badRequest(w, errors.New("unit_number is required"))
		return
	}

	id, err := handler.insertTrailer(r, request)
	if err != nil {
		serverError(w, err)
		return
	}

	handler.show(w, r, id)
}

func (handler *TrailersHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := requireRole(w, r, middleware.RoleAdmin, middleware.RoleFleetManager); !ok {
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request trailerPatchRequest
	if err := decodeJSON(r, &request); err != nil {
		badRequest(w, err)
		return
	}

	updatedID, err := handler.patchTrailer(r, id, request)
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

func (handler *TrailersHandler) insertTrailer(r *http.Request, request trailerCreateRequest) (string, error) {
	params, err := buildTrailerCreateParams(request)
	if err != nil {
		return "", err
	}

	row := handler.database.QueryRowContext(r.Context(), db.CreateTrailerQuery,
		params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7], params[8],
	)

	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (handler *TrailersHandler) patchTrailer(r *http.Request, id string, request trailerPatchRequest) (string, error) {
	params, err := buildTrailerPatchParams(request, id)
	if err != nil {
		return "", err
	}

	row := handler.database.QueryRowContext(r.Context(), db.UpdateTrailerQuery,
		params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7], params[8], params[9],
	)

	var updatedID string
	if err := row.Scan(&updatedID); err != nil {
		return "", err
	}

	return updatedID, nil
}

func buildTrailerCreateParams(request trailerCreateRequest) ([]any, error) {
	return []any{
		strings.TrimSpace(request.UnitNumber),
		nullString(request.Vin),
		nullString(request.PlateNumber),
		nullInt(request.Year),
		nullString(request.Make),
		nullString(request.UsageType),
		nullString(request.Location),
		nullString(request.Availability),
		nullString(request.Notes),
	}, nil
}

func buildTrailerPatchParams(request trailerPatchRequest, id string) ([]any, error) {
	return []any{
		nullString(request.UnitNumber),
		nullString(request.Vin),
		nullString(request.PlateNumber),
		nullInt(request.Year),
		nullString(request.Make),
		nullString(request.UsageType),
		nullString(request.Location),
		nullString(request.Availability),
		nullString(request.Notes),
		id,
	}, nil
}

func scanTrailer(scanner rowScanner) (models.Trailer, error) {
	var trailer models.Trailer
	var vin, plateNumber, makeValue, usageType, location, availability, notes sql.NullString
	var year sql.NullInt32

	if err := scanner.Scan(
		&trailer.ID,
		&trailer.UnitNumber,
		&vin,
		&plateNumber,
		&year,
		&makeValue,
		&usageType,
		&location,
		&availability,
		&notes,
		&trailer.CreatedAt,
		&trailer.UpdatedAt,
	); err != nil {
		return models.Trailer{}, err
	}

	trailer.Vin = stringPtrFromNull(vin)
	trailer.PlateNumber = stringPtrFromNull(plateNumber)
	trailer.Year = intPtrFromNull(year)
	trailer.Make = stringPtrFromNull(makeValue)
	trailer.UsageType = stringPtrFromNull(usageType)
	trailer.Location = stringPtrFromNull(location)
	trailer.Availability = stringPtrFromNull(availability)
	trailer.Notes = stringPtrFromNull(notes)

	return trailer, nil
}
