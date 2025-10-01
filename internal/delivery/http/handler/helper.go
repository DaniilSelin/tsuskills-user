package handler

import (
	"context"
	"net/http"
	"strings"

	"tsuskills-user/internal/delivery/dto"
	"tsuskills-user/internal/delivery/validator"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// parseAndValidateJWTToken извлекает извлекает {token} из заголовков, оборачивает в DTO и валидирует.
// При ошибке возвращает HTTP‑код и сообщение, чтобы handler мог сразу ответить.
func (h *Handler) parseAndValidateJWTToken(
	ctx context.Context,
	r *http.Request,
	operation string,
) (*dto.JWTToken, int, string) {
	token := r.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer")
	if len(splitToken) != 2 {
		h.log.Warn(ctx, operation+"token not provided")
		return nil, http.StatusBadRequest, "token not provided"
	}
	jwtToken := dto.JWTToken{Token: splitToken[1]}
	if err := validator.ValidateStruct(jwtToken); err != nil {
		h.log.Warn(ctx, operation+"token validation failed", zap.Error(err))
		return nil, http.StatusBadRequest, err.Error()
	}
	return &jwtToken, 0, ""
}

func (h Handler) parseAndValidateIdUser(
	ctx context.Context,
	r *http.Request,
	operation string,
) (*uuid.UUID, int, string) {
	vars := mux.Vars(r)
	userID, ok := vars["id"]
	if !ok {
		h.log.Warn(ctx, operation+"id not provided")
		return nil, http.StatusBadRequest, "id not provided"
	}
	uuidID, err := uuid.FromBytes([]byte(userID))
	if err != nil {
		h.log.Warn(ctx, operation+"id ")
		return nil, http.StatusBadRequest, "id not provided"
	}
	id := dto.UserID{ID: uuidID} // место для валидации uuid, хеша uuid
	if err := validator.ValidateStruct(id); err != nil {
		h.log.Warn(ctx, operation+"user id validation failed", zap.Error(err))
		return nil, http.StatusBadRequest, err.Error()
	}
	return &id.ID, 0, ""
}
