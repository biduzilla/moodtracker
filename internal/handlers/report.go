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

type reportHandler struct {
	report       services.ReportService
	errorHandler e.ErrorHandlerInterface
}

type ReportHandler interface {
	GetMonthlyReport(w http.ResponseWriter, r *http.Request)
	GetTagReport(w http.ResponseWriter, r *http.Request)
	GetMoodReport(w http.ResponseWriter, r *http.Request)
}

func NewReportHandler(
	report services.ReportService,
	errorHandler e.ErrorHandlerInterface,
) *reportHandler {
	return &reportHandler{
		report:       report,
		errorHandler: errorHandler,
	}
}

func (h *reportHandler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	year := utils.ReadIntParam(r, "year", time.Now().Year(), v)
	month := utils.ReadIntParam(r, "month", int(time.Now().Month()), v)

	if !v.Valid() {
		h.errorHandler.HandlerError(w, r, e.ErrInvalidData, v)
		return
	}

	user := contexts.ContextGetUser(r)
	report, err := h.report.GetMonthlyReport(year, month, user.ID)
	if err != nil {
		h.errorHandler.HandlerError(w, r, err, v)
		return
	}

	respond(w, r, http.StatusOK, utils.Envelope{utils.GetTypeName(report): report}, nil, h.errorHandler)
}

func (h *reportHandler) GetTagReport(w http.ResponseWriter, r *http.Request) {
	tag := utils.ReadStringParam(r, "tag", "")
	user := contexts.ContextGetUser(r)

	tagReport, err := h.report.GetTagReport(tag, user.ID)
	if err != nil {
		h.errorHandler.HandlerError(w, r, err, nil)
		return
	}

	respond(w, r, http.StatusOK, utils.Envelope{utils.GetTypeName(tagReport): tagReport}, nil, h.errorHandler)
}

func (h *reportHandler) GetMoodReport(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	moodLabel := utils.ReadIntParam(r, "mood_label", 1, v)

	if !v.Valid() {
		h.errorHandler.HandlerError(w, r, e.ErrInvalidData, v)
		return
	}

	user := contexts.ContextGetUser(r)

	moodReport, err := h.report.GetMoodReport(models.MoodLabel(moodLabel), user.ID)
	if err != nil {
		h.errorHandler.HandlerError(w, r, err, nil)
		return
	}

	respond(w, r, http.StatusOK, utils.Envelope{utils.GetTypeName(moodReport): moodReport}, nil, h.errorHandler)
}
