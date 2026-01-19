package handlers

import (
	"database/sql"
	"moodtracker/internal/config"
	"moodtracker/internal/jsonlog"
	"moodtracker/internal/services"
	"moodtracker/utils"
	"moodtracker/utils/errors"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	User    UserHandlerInterface
	Auth    AuthHandlerInterface
	Service *services.Services
	Daylog  DaylogHandler
	Tag     TagHandler
}

func NewHandler(
	db *sql.DB,
	errRsp errors.ErrorHandlerInterface,
	config config.Config,
	logger jsonlog.Logger,
) *Handler {
	s := services.NewServices(logger, db, config)

	return &Handler{
		Service: s,
		User:    NewUserHandler(s.User, errRsp),
		Auth:    NewAuthHandler(s.Auth, errRsp),
		Daylog:  NewDaylogHandler(s.Daylog, errRsp),
		Tag:     NewTagHandler(s.Tag, errRsp),
	}
}

func parseIntID(
	w http.ResponseWriter,
	r *http.Request,
	errRsp errors.ErrorHandlerInterface,
) (int64, bool) {
	id, err := utils.ReadIntPathVariable(r, "id")
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return 0, false
	}
	return id, true
}

func parseUUID(
	w http.ResponseWriter,
	r *http.Request,
	errRsp errors.ErrorHandlerInterface,
) (uuid.UUID, bool) {

	id, err := utils.ReadStringPathVariable(r, "id")
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return uuid.Nil, false
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return uuid.Nil, false
	}

	return uid, true
}

func respond(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data utils.Envelope,
	headers http.Header,
	errRsp errors.ErrorHandlerInterface,
) {
	err := utils.WriteJSON(w, status, data, headers)
	if err != nil {
		errRsp.ServerErrorResponse(w, r, err)
	}
}
