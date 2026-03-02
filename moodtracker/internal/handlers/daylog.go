package handlers

import (
	"moodtracker/internal/contexts"
	"moodtracker/internal/models"
	"moodtracker/internal/services"
	"moodtracker/utils"
	e "moodtracker/utils/errors"
	"moodtracker/utils/validator"
	"net/http"
	"time"
)

type daylogHandlers struct {
	daylog services.DaylogServices
	errRsp e.ErrorHandlerInterface
	GenericHandlerInterface[models.Daylog, models.DaylogDTO]
}

func NewDaylogHandler(
	daylog services.DaylogServices,
	errRsp e.ErrorHandlerInterface,
) *daylogHandlers {
	return &daylogHandlers{
		daylog:                  daylog,
		errRsp:                  errRsp,
		GenericHandlerInterface: NewGenericHandler(daylog, errRsp),
	}
}

type DaylogHandler interface {
	GetAllByYear(w http.ResponseWriter, r *http.Request)
	GenericHandlerInterface[
		models.Daylog,
		models.DaylogDTO,
	]
}

func (h *daylogHandlers) GetAllByYear(w http.ResponseWriter, r *http.Request) {
	v := validator.New()
	year := utils.ReadIntParam(r, "year", time.Now().Year(), v)

	if !v.Valid() {
		h.errRsp.HandlerError(w, r, e.ErrInvalidData, v)
		return
	}

	user := contexts.ContextGetUser(r)
	datas, err := h.daylog.GetAllByYear(year, user.ID)
	if err != nil {
		h.errRsp.HandlerError(w, r, err, v)
		return
	}

	dtos := make([]*models.DaylogDTO, 0, len(datas))
	for _, m := range datas {
		dto := m.ToDTO()
		dtos = append(dtos, dto)
	}

	respond(w, r, http.StatusOK, utils.Envelope{"day_logs": dtos}, nil, h.errRsp)
}
