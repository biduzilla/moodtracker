package handlers

import (
	"moodtracker/internal/contexts"
	"moodtracker/internal/models"
	"moodtracker/internal/models/filters"
	"moodtracker/internal/services"
	"moodtracker/utils"
	e "moodtracker/utils/errors"
	"moodtracker/utils/validator"
	"net/http"
)

type tagHandler struct {
	tag    services.TagService
	errRsp e.ErrorHandlerInterface
	GenericHandlerInterface[models.Tag, models.TagDTO]
}

func NewTagHandler(
	tag services.TagService,
	errRsp e.ErrorHandlerInterface,
) *tagHandler {
	return &tagHandler{
		tag:                     tag,
		errRsp:                  errRsp,
		GenericHandlerInterface: NewGenericHandler(tag, errRsp),
	}
}

type TagHandler interface {
	GetAllByUserID(w http.ResponseWriter, r *http.Request)
	GenericHandlerInterface[
		models.Tag,
		models.TagDTO,
	]
}

func (h *tagHandler) GetAllByUserID(w http.ResponseWriter, r *http.Request) {
	var input struct {
		title,
		author string
		filters.Filters
	}

	v := validator.New()
	qs := r.URL.Query()
	input.Filters.Page = utils.ReadInt(r, "page", 1, v)
	input.Filters.PageSize = utils.ReadInt(r, "page_size", 20, v)
	input.Filters.Sort = utils.ReadString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "-id", "-name"}

	user := contexts.ContextGetUser(r)
	tags, metadata, err := h.tag.GetAllByUserID(user.ID, input.Filters)
	if err != nil {
		h.errRsp.HandlerError(w, r, err, v)
		return
	}

	dtos := make([]*models.TagDTO, 0, len(tags))
	for _, t := range tags {
		dtos = append(dtos, t.ToDTO())
	}

	respond(w, r, http.StatusOK, utils.Envelope{"tags": dtos, "metadata": metadata}, nil, h.errRsp)
}
