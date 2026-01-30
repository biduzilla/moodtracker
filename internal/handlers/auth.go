package handlers

import (
	"moodtracker/internal/services"
	"moodtracker/utils"
	"moodtracker/utils/errors"
	"moodtracker/utils/validator"
	"net/http"
)

type AuthHandler struct {
	auth         services.AuthServiceInterface
	errorHandler errors.ErrorHandlerInterface
}

type AuthHandlerInterface interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
}

func NewAuthHandler(authService services.AuthServiceInterface, errResp errors.ErrorHandlerInterface) *AuthHandler {
	return &AuthHandler{
		auth:         authService,
		errorHandler: errResp,
	}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.errorHandler.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	token, err := h.auth.Login(v, input.Email, input.Password)

	if err != nil {
		h.errorHandler.HandlerError(w, r, err, v)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"authentication_token": token}, nil)
	if err != nil {
		h.errorHandler.ServerErrorResponse(w, r, err)
	}
}
