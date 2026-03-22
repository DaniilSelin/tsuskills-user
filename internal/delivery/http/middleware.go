package http

import (
	"context"
	"fmt"
	httpBase "net/http"
	"time"

	"tsuskills-user/internal/domain"
	"tsuskills-user/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func RequestIDMiddleware(next httpBase.Handler) httpBase.Handler {
	return httpBase.HandlerFunc(func(w httpBase.ResponseWriter, r *httpBase.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), logger.RequestID, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CORSMiddleware(next httpBase.Handler) httpBase.Handler {
	return httpBase.HandlerFunc(func(w httpBase.ResponseWriter, r *httpBase.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == httpBase.MethodOptions {
			w.WriteHeader(httpBase.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(log logger.Logger) func(httpBase.Handler) httpBase.Handler {
	return func(next httpBase.Handler) httpBase.Handler {
		return httpBase.HandlerFunc(func(w httpBase.ResponseWriter, r *httpBase.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, statusCode: httpBase.StatusOK}

			next.ServeHTTP(wrapped, r)

			log.Info(r.Context(), "HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status_code", wrapped.statusCode),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

func RecoveryMiddleware(log logger.Logger) func(httpBase.Handler) httpBase.Handler {
	return func(next httpBase.Handler) httpBase.Handler {
		return httpBase.HandlerFunc(func(w httpBase.ResponseWriter, r *httpBase.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error(r.Context(), "Panic recovered", zap.Any("error", err))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(httpBase.StatusInternalServerError)
					resp := fmt.Sprintf(
						`{"error":"Internal Server Error","code":"%s","message":"An unexpected error occurred"}`,
						domain.CodeInternal,
					)
					w.Write([]byte(resp))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type responseWriter struct {
	httpBase.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
