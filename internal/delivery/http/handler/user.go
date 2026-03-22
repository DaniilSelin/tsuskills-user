package handler

import (
	"net/http"

	"tsuskills-user/internal/delivery/dto"
	"tsuskills-user/internal/delivery/mapper"
	"tsuskills-user/internal/domain"

	"go.uber.org/zap"
)

// HandleGetUser обрабатывает GET /api/v1/users/{id}
func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleGetUser"

	id, ok := h.extractUUIDParam(r, "id")
	if !ok {
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, "Invalid or missing user ID")
		return
	}

	user, email, svcCode := h.userSV.GetUser(ctx, id)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(ctx, op+": success", zap.String("user_id", id.String()))

	resp := mapper.UserToResponse(user, email)
	h.writeJSON(ctx, w, http.StatusOK, resp)
}

// HandleGetMe обрабатывает GET /api/v1/users/me
// Достаёт user_id из Bearer token и возвращает профиль.
func (h *Handler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleGetMe"

	token, ok := h.extractBearerToken(r)
	if !ok {
		h.writeError(ctx, w, http.StatusUnauthorized, domain.CodeUnauthorized, "Bearer token required")
		return
	}

	userID, svcCode := h.userSV.Auth(ctx, token)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	user, email, svcCode := h.userSV.GetUser(ctx, userID)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	resp := mapper.UserToResponse(user, email)
	h.writeJSON(ctx, w, http.StatusOK, resp)
}

// HandleUpdateUser обрабатывает PUT /api/v1/users/{id}
func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleUpdateUser"

	id, ok := h.extractUUIDParam(r, "id")
	if !ok {
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, "Invalid or missing user ID")
		return
	}

	var req dto.UpdateUserRequest
	if !h.decodeAndValidate(ctx, w, r, &req) {
		return
	}

	// получаем текущего пользователя
	user, _, svcCode := h.userSV.GetUser(ctx, id)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	// обновляем только переданные поля
	if req.Name != "" {
		user.Name = req.Name
	}

	svcCode = h.userSV.UpdateUser(ctx, user, req.NewPassword)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(ctx, op+": success", zap.String("user_id", id.String()))

	h.writeJSON(ctx, w, http.StatusOK, map[string]string{"message": "user updated"})
}

// HandleDeleteUser обрабатывает DELETE /api/v1/users/{id}
func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleDeleteUser"

	id, ok := h.extractUUIDParam(r, "id")
	if !ok {
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, "Invalid or missing user ID")
		return
	}

	svcCode := h.userSV.DeleteUser(ctx, id)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(ctx, op+": success", zap.String("user_id", id.String()))

	h.writeJSON(ctx, w, http.StatusOK, map[string]string{"message": "user deleted"})
}
