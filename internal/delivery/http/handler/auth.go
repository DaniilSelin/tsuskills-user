package handler

import (
	"net/http"

	"tsuskills-user/internal/delivery/dto"
	"tsuskills-user/internal/delivery/mapper"
	"tsuskills-user/internal/domain"

	"go.uber.org/zap"
)

// HandleRegister обрабатывает POST /api/v1/users/register
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleRegister"

	var req dto.RegistrationRequest
	if !h.decodeAndValidate(ctx, w, r, &req) {
		return
	}

	h.log.Info(ctx, op+": received", zap.String("email", req.Email))

	domainReq := mapper.RegReqToDomain(req)

	userID, accessToken, refreshToken, svcCode := h.userSV.Register(ctx, domainReq)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(ctx, op+": success", zap.String("user_id", userID.String()))

	h.writeJSON(ctx, w, http.StatusCreated, dto.AuthResponse{
		UserID:       userID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// HandleLogin обрабатывает POST /api/v1/users/login
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleLogin"

	var req dto.LoginRequest
	if !h.decodeAndValidate(ctx, w, r, &req) {
		return
	}

	h.log.Info(ctx, op+": received", zap.String("email", req.Email))

	domainReq := mapper.LoginReqToDomain(req)

	accessToken, refreshToken, svcCode := h.userSV.Login(ctx, domainReq)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(ctx, op+": success", zap.String("email", req.Email))

	h.writeJSON(ctx, w, http.StatusOK, dto.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// HandleRefreshToken обрабатывает POST /api/v1/users/refresh
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleRefreshToken"

	var req dto.RefreshRequest
	if !h.decodeAndValidate(ctx, w, r, &req) {
		return
	}

	newAccessToken, svcCode := h.userSV.RefreshToken(ctx, req.RefreshToken)
	if svcCode != domain.CodeOK {
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(ctx, op+": success")

	h.writeJSON(ctx, w, http.StatusOK, map[string]string{
		"access_token": newAccessToken,
	})
}

// HandleAuth обрабатывает GET /api/v1/users/auth
// Проверяет Bearer token из заголовка и возвращает user_id.
func (h *Handler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const op = "HandleAuth"

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

	h.log.Info(ctx, op+": success", zap.String("user_id", userID.String()))

	h.writeJSON(ctx, w, http.StatusOK, map[string]string{
		"user_id": userID.String(),
	})
}
