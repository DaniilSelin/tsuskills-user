package http

import (
	httpBase "net/http"
	"tsuskills-user/internal/logger"

	"github.com/gorilla/mux"
)

type IHandler interface {
	// Auth endpoints
	HandleRegister(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleLogin(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleRefreshToken(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleAuth(w httpBase.ResponseWriter, r *httpBase.Request)

	// User endpoints
	HandleGetMe(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleGetUser(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleUpdateUser(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleDeleteUser(w httpBase.ResponseWriter, r *httpBase.Request)
}

func NewRouter(h IHandler, log logger.Logger) *mux.Router {
	r := mux.NewRouter()

	r.Use(RequestIDMiddleware)
	r.Use(CORSMiddleware)
	r.Use(LoggingMiddleware(log))
	r.Use(RecoveryMiddleware(log))

	api := r.PathPrefix("/api/v1/users").Subrouter()

	// Auth — публичные эндпоинты
	api.HandleFunc("/register", h.HandleRegister).Methods(httpBase.MethodPost, httpBase.MethodOptions)
	api.HandleFunc("/login", h.HandleLogin).Methods(httpBase.MethodPost, httpBase.MethodOptions)
	api.HandleFunc("/refresh", h.HandleRefreshToken).Methods(httpBase.MethodPost, httpBase.MethodOptions)
	api.HandleFunc("/auth", h.HandleAuth).Methods(httpBase.MethodGet, httpBase.MethodOptions)

	// User profile — требуют авторизации (проверка на уровне handler)
	api.HandleFunc("/me", h.HandleGetMe).Methods(httpBase.MethodGet, httpBase.MethodOptions)
	api.HandleFunc("/{id}", h.HandleGetUser).Methods(httpBase.MethodGet, httpBase.MethodOptions)
	api.HandleFunc("/{id}", h.HandleUpdateUser).Methods(httpBase.MethodPut, httpBase.MethodOptions)
	api.HandleFunc("/{id}", h.HandleDeleteUser).Methods(httpBase.MethodDelete, httpBase.MethodOptions)

	// Health check
	r.HandleFunc("/health", func(w httpBase.ResponseWriter, r *httpBase.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpBase.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods(httpBase.MethodGet)

	return r
}
