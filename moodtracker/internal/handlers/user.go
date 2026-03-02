package handlers

import (
	"moodtracker/internal/models"
	"moodtracker/internal/services"
	"moodtracker/utils"
	e "moodtracker/utils/errors"
	"moodtracker/utils/validator"
	"net/http"
)

type UserHandler struct {
	user   services.UserService
	errRsp e.ErrorHandlerInterface
}

type UserHandlerInterface interface {
	ActivateUserHandler(w http.ResponseWriter, r *http.Request)
	CreateUserHandler(w http.ResponseWriter, r *http.Request)
}

func NewUserHandler(
	user services.UserService,
	errRsp e.ErrorHandlerInterface,
) *UserHandler {
	return &UserHandler{
		user:   user,
		errRsp: errRsp,
	}
}

func (h *UserHandler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Cod   int    `json:"cod"`
		Email string `json:"email"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.errRsp.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	user, err := h.user.ActivateUser(
		input.Cod,
		input.Email,
		v,
	)
	if err != nil {
		h.errRsp.HandlerError(w, r, err, v)
		return
	}

	respond(
		w,
		r,
		http.StatusOK,
		utils.Envelope{"user": user.ToDTO()},
		nil,
		h.errRsp,
	)
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userDTO models.UserSaveDTO
	if err := utils.ReadJSON(w, r, &userDTO); err != nil {
		h.errRsp.BadRequestResponse(w, r, err)
		return
	}

	user, err := userDTO.ToModel()
	if err != nil {
		h.errRsp.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	err = h.user.Save(user, v)
	if err != nil {
		h.errRsp.HandlerError(w, r, err, v)
		return
	}

	respond(
		w,
		r,
		http.StatusCreated,
		utils.Envelope{"user": user.ToDTO()},
		nil,
		h.errRsp,
	)
}
