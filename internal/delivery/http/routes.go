package http

import (
	httpBase "net/http"
	"tsuskills-user/internal/logger"

	"github.com/gorilla/mux"
)

type IHanlder interface {
	HandleCreateUser(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleLogin(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleRefreshToken(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleRegister(w httpBase.ResponseWriter, r *httpBase.Request)
	HandleAuth(w httpBase.ResponseWriter, r *httpBase.Request)
}

func NewRouter(h IHanlder, log logger.Logger) *mux.Router {
	r := mux.NewRouter()

	// Добавляем middleware
	r.Use(LoggingMiddleware(log))
	r.Use(RecoveryMiddleware(log))

	api := r.PathPrefix("/api/v1/users").Subrouter()

	api.HandleFunc("", h.HandleRegister).Methods(httpBase.MethodPost)
	api.HandleFunc("/login", h.HandleLogin).Methods(httpBase.MethodPost)
	api.HandleFunc("/refresh", h.HandleRefreshToken).Methods(httpBase.MethodPost)

	api.HandleFunc("/auth", h.HandleAuth).Methods(httpBase.MethodGet)

	return r
}
