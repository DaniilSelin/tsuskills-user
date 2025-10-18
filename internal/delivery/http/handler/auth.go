package handler

import (
	"encoding/json"
	"net/http"
	"tsuskills-user/internal/delivery/dto"
	"tsuskills-user/internal/delivery/mapper"
	"tsuskills-user/internal/delivery/validator"
	"tsuskills-user/internal/domain"

	"go.uber.org/zap"
)

func (h *Handler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := h.nameFunc()

	h.log.Info(
		ctx,
		op+": invoked",
	)

	token, code, msg := h.parseAndValidateJWTToken(ctx, r, op)
	if code != 0 {
		h.writeError(ctx, w, code, domain.CodeInvalidRequestBody, msg)
		return
	}

	h.log.Info(
		ctx,
		op+"received request",
		zap.Any("payload", token),
	)

	userID, svcCode := h.userSV.Auth(ctx, token.Token)
	if svcCode != domain.CodeOK {
		h.log.Error(
			ctx,
			op+": auth failed",
			zap.String("error_code", string(svcCode)),
		)
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(
		ctx,
		op+": token auth successful",
		zap.Int("userID", userID),
	)

	h.writeJSON(ctx, w, http.StatusOK, map[string]int{"user_id": userID})
}

func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := h.nameFunc()

	h.log.Info(
		ctx,
		op+": invoked",
	)

	token, code, msg := h.parseAndValidateJWTToken(ctx, r, op)
	if code != 0 {
		h.writeError(ctx, w, code, domain.CodeInvalidRequestBody, msg)
		return
	}

	h.log.Info(
		ctx,
		op+"received request",
		zap.Any("payload", token),
	)

	newToken, svcCode := h.userSV.RefreshToken(ctx, token.Token)
	if svcCode != domain.CodeOK {
		h.log.Error(
			ctx,
			op+": refresh token failed",
			zap.String("error_code", string(svcCode)),
		)
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(
		ctx,
		op+"received request",
		zap.Any("payload", newToken),
	)

	h.writeJSON(ctx, w, http.StatusOK, map[string]string{"access token": newToken})
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := h.nameFunc()

	h.log.Info(
		ctx,
		op+": invoked",
	)

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(
			ctx,
			op+"failed to decode JSON",
			zap.Error(err),
		)
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, "Invalid JSON")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		h.log.Warn(
			ctx,
			op+"validation failed",
			zap.Any("errors", err),
		)
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, err.Error())
		return
	}

	logReq := mapper.LoginReqToLoginModel(req)

	h.log.Info(
		ctx,
		op+"received request",
		zap.Any("payload", req),
	)

	token, svcCode := h.userSV.Login(ctx, logReq)
	if svcCode != domain.CodeOK {
		h.log.Warn(
			ctx,
			op+": login failed",
			zap.String("error_code", string(svcCode)),
		)
		h.handleServiceError(ctx, w, svcCode, "Login failed")
		return
	}

	h.log.Info(
		ctx,
		op+": login successfully",
		zap.String("token", token),
	)
	h.writeJSON(ctx, w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := h.nameFunc()

	h.log.Info(
		ctx,
		op+": invoked",
	)

	var req dto.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(
			ctx,
			op+"failed to decode JSON",
			zap.Error(err),
		)
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, "Invalid JSON")
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		h.log.Warn(
			ctx,
			op+"validation failed",
			zap.Any("errors", err),
		)
		h.writeError(ctx, w, http.StatusBadRequest, domain.CodeInvalidRequestBody, err.Error())
		return
	}

	h.log.Info(
		ctx,
		op+"received request",
		zap.Any("payload", req),
	)

	user := mapper.RegRequestToUser(req)

	userID, token, svcCode := h.userSV.Register(ctx, user)
	if svcCode != domain.CodeOK {
		h.log.Warn(
			ctx,
			op+": user registration failed",
			zap.String("error_code", string(svcCode)),
		)
		h.handleServiceError(ctx, w, svcCode, "Registration failed")
		return
	}

	h.log.Info(
		ctx,
		op+": registered successfully",
		zap.Any("userID", userID),
		zap.String("token", token),
	)
	h.writeJSON(ctx, w, http.StatusOK, map[string]interface{}{"user_id": userID, "token": token})
}
