package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"tsuskills-user/internal/delivery/validator"
	"tsuskills-user/internal/domain"
	"tsuskills-user/internal/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type IUserService interface {
	Register(ctx context.Context, req domain.RegistrationRequest) (uuid.UUID, string, string, domain.ErrorCode)
	Login(ctx context.Context, req domain.LoginRequest) (string, string, domain.ErrorCode)
	Auth(ctx context.Context, token string) (uuid.UUID, domain.ErrorCode)
	RefreshToken(ctx context.Context, refreshToken string) (string, domain.ErrorCode)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.User, *domain.Email, domain.ErrorCode)
	UpdateUser(ctx context.Context, user *domain.User, newPassword string) domain.ErrorCode
	DeleteUser(ctx context.Context, id uuid.UUID) domain.ErrorCode
}

type Handler struct {
	userSV IUserService
	log    logger.Logger
}

func NewHandler(svc IUserService, l logger.Logger) *Handler {
	return &Handler{userSV: svc, log: l}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

func (h *Handler) writeJSON(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error(ctx, "Failed to encode JSON", zap.Error(err))
	}
}

func (h *Handler) writeError(ctx context.Context, w http.ResponseWriter, statusCode int, code domain.ErrorCode, message string) {
	h.writeJSON(ctx, w, statusCode, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    string(code),
		Message: message,
	})
}

func (h *Handler) handleServiceError(ctx context.Context, w http.ResponseWriter, code domain.ErrorCode, operation string) {
	switch code {
	case domain.CodeOK:
		return
	case domain.CodeNotFound:
		h.writeError(ctx, w, http.StatusNotFound, code, "Resource not found")
	case domain.CodeConflict:
		h.writeError(ctx, w, http.StatusConflict, code, "Resource already exists")
	case domain.CodeUnauthorized:
		h.writeError(ctx, w, http.StatusUnauthorized, code, "Unauthorized")
	case domain.CodeInvalidCredentials:
		h.writeError(ctx, w, http.StatusUnauthorized, code, "Invalid email or password")
	case domain.CodeForbidden:
		h.writeError(ctx, w, http.StatusForbidden, code, "Forbidden")
	case domain.CodeInvalidRequestBody:
		h.writeError(ctx, w, http.StatusBadRequest, code, "Invalid request")
	default:
		h.log.Error(ctx, operation+": internal error", zap.String("code", string(code)))
		h.writeError(ctx, w, http.StatusInternalServerError, domain.CodeInternal, "Internal server error")
	}
}

// decodeAndValidate декодирует JSON body и валидирует структуру.
// При ошибке сам пишет ответ в w и возвращает false.
func (h *Handler) decodeAndValidate(ctx context.Context, w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, "Invalid JSON")
		return false
	}
	if err := validator.ValidateStruct(dst); err != nil {
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, err.Error())
		return false
	}
	return true
}

// extractBearerToken достаёт токен из заголовка Authorization: Bearer <token>
func (h *Handler) extractBearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}

// extractUUIDParam достаёт UUID параметр из URL path.
func (h *Handler) extractUUIDParam(r *http.Request, name string) (uuid.UUID, bool) {
	vars := mux.Vars(r)
	raw, ok := vars[name]
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
