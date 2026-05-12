package middleware

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (writer *loggingResponseWriter) WriteHeader(statusCode int) {
	writer.statusCode = statusCode
	writer.ResponseWriter.WriteHeader(statusCode)
}

func (writer *loggingResponseWriter) Write(data []byte) (int, error) {
	if writer.statusCode == 0 {
		writer.statusCode = http.StatusOK
	}
	n, err := writer.ResponseWriter.Write(data)
	writer.bytes += n
	return n, err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		log.Printf("%s %s %d %dB %s", r.Method, r.URL.Path, writer.statusCode, writer.bytes, time.Since(start).Round(time.Millisecond))
	})
}
