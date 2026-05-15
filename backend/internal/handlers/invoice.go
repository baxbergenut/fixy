package handlers

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"errors"

	"fixy-backend/internal/middleware"
	"fixy-backend/internal/services"
)

type InvoiceHandler struct {
	groqToken string
}

func NewInvoiceHandler(groqToken string) http.Handler {
	return &InvoiceHandler{groqToken: groqToken}
}

func (handler *InvoiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/invoice/parse" {
		notFound(w)
		return
	}

	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	if _, ok := requireRole(w, r, middleware.RoleAdmin, middleware.RoleAccountant, middleware.RoleFleetManager); !ok {
		return
	}

	handler.parse(w, r)
}

func (handler *InvoiceHandler) parse(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(handler.groqToken) == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "GROQ_TOKEN is not configured"})
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		badRequest(w, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		badRequest(w, err)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		serverError(w, err)
		return
	}

	mimeType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if mimeType == "" || mimeType == "application/octet-stream" {
		if extensionMimeType := mime.TypeByExtension(filepath.Ext(header.Filename)); extensionMimeType != "" {
			mimeType = extensionMimeType
		} else {
			mimeType = http.DetectContentType(data)
		}
	}

	result, err := services.ParseInvoice(r.Context(), handler.groqToken, mimeType, data)
	if err != nil {
		var upstreamErr *services.UpstreamStatusError
		if errors.As(err, &upstreamErr) {
			writeJSON(w, upstreamErr.StatusCode, map[string]string{"error": upstreamErr.Message})
			return
		}
		serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
