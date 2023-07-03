package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-web-api/pkg/ctxfields"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("unable to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		started := time.Now()
		wrapped := NewResponseWriterWrapper(w)
		next.ServeHTTP(wrapped, r)

		log.Info().
			Str("url", r.URL.String()).
			Str("session", r.Header.Get("authorization")).
			Str("body", string(body)).
			Str("method", r.Method).
			Int("status", wrapped.StatusCode).
			Str("ip", ctxfields.ExtractRequestIP(r.Context())).
			Str("duration", time.Since(started).String()).
			Msg("incoming request")
	})
}

type ResponseWriterWrapper struct {
	StatusCode int
	writer     http.ResponseWriter
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{writer: w}
}

func (w *ResponseWriterWrapper) Header() http.Header {
	return w.writer.Header()
}

func (w *ResponseWriterWrapper) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.writer.WriteHeader(statusCode)
}
