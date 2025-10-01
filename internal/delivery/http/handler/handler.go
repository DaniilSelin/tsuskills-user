package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"

	"tsuskills-user/internal/domain"
	"tsuskills-user/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// IUserService определяет интерфейс для работы с кошельками.
type IUserService interface {
	CreateUser(ctx context.Context, user *domain.User) (int, domain.ErrorCode)
	GetUser(ctx context.Context, id *uuid.UUID) (domain.User, domain.Email, domain.ErrorCode)
	UpdateUser(ctx context.Context, user domain.User) domain.ErrorCode
	DeleteUser(ctx context.Context, id int) error
	Login(ctx context.Context, login domain.LoginRequest) (string, domain.ErrorCode)
	Auth(ctx context.Context, token string) (int, domain.ErrorCode)
	RefreshToken(ctx context.Context, token string) (string, domain.ErrorCode)
	Register(ctx context.Context, registry domain.RegistrationRequest) (uuid.UUID, string, domain.ErrorCode)
}

// Handler - HTTP обработчик для API.
// Содержит зависимости на сервисы транзакций и кошельков, а также логгер.
type Handler struct {
	userSV IUserService
	log    logger.Logger
}

// NewHandler создает новый экземпляр HTTP обработчика.
// Принимает сервисы транзакций и кошельков, а также логгер.
// Возвращает указатель на Handler.
func NewHandler(ts IUserService, l logger.Logger) *Handler {
	return &Handler{
		userSV: ts,
		log:    l,
	}
}

// ErrorResponse представляет структуру ответа с ошибкой.
// Используется для стандартизации формата ошибок API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// writeJSON записывает JSON ответ в http.ResponseWriter.
// Устанавливает Content-Type header и кодирует данные в JSON.
// При ошибке кодирования логирует её.
func (h *Handler) writeJSON(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error(ctx, "Failed to encode JSON response", zap.Error(err))
	}
}

// writeError записывает ошибку в формате JSON в http.ResponseWriter.
// Создает ErrorResponse с указанным статус кодом, кодом ошибки и сообщением.
func (h *Handler) writeError(ctx context.Context, w http.ResponseWriter, statusCode int, code domain.ErrorCode, message string) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    string(code),
		Message: message,
	}
	h.writeJSON(ctx, w, statusCode, response)
}

// handleServiceError маппит коды ошибок домена в соответствующие HTTP статус коды и сообщения.
// Принимает код ошибки домена и название операции для логирования.
func (h *Handler) handleServiceError(ctx context.Context, w http.ResponseWriter, code domain.ErrorCode, operation string) {
	switch code {
	case domain.CodeOK:
		return
	case domain.CodeInvalidLimit:
		h.log.Warn(ctx, operation+": invalid limit")
		h.writeError(ctx, w, http.StatusBadRequest, code, "Invalid limit parameter")
	case domain.CodeInternal:
		h.log.Error(ctx, operation+": internal error")
		h.writeError(ctx, w, http.StatusInternalServerError, code, "Internal server error")
	default:
		h.log.Error(ctx, operation+": unknown error code", zap.String("code", string(code)))
		h.writeError(ctx, w, http.StatusInternalServerError, domain.CodeInternal, "Internal server error")
	}
}

// currentFunction возвращает имя функции, в которой вызывается
// при ошибке возвращает Unknown
// ИСПОЛЬЗУЕТСЯ ТОЛЬКО В
func (h Handler) nameFunc() string {
	callers, _, _, ok := runtime.Caller(3) // Сама Caller тоже в стеке, 0 скипаем
	if !ok {
		return "Unknown"
	}
	return runtime.FuncForPC(callers).Name()
}
