package handler

import (
	"net/http"

	"tsuskills-user/internal/delivery/mapper"
	"tsuskills-user/internal/domain"

	"go.uber.org/zap"
)

func (h *Handler) HandleReadUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	op := h.nameFunc()

	h.log.Info(
		ctx,
		op+": invoked",
	)

	id, code, msg := h.parseAndValidateIdUser(ctx, r, op)
	if code != 0 {
		h.writeError(ctx, w, code, domain.CodeInvalidRequestBody, msg)
		return
	}

	user, email, svcCode := h.userSV.GetUser(ctx, id)
	if svcCode != domain.CodeOK {
		h.log.Error(
			ctx,
			op+": failed read user",
			zap.String("error_code", string(svcCode)),
		)
		h.handleServiceError(ctx, w, svcCode, op)
		return
	}

	h.log.Info(
		ctx,
		"User successfully retrieved",
		zap.Any("userID", id),
	)

	resp := mapper.UserToRespUser(user, email)

	h.writeJSON(ctx, w, http.StatusOK, resp)
}
