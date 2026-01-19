package models

import (
	"moodtracker/utils"
	"moodtracker/utils/validator"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MoodLabel int

const (
	MOOD_RUIM MoodLabel = iota + 1
	MOOD_MEDIO
	MOOD_BOM
)

type Daylog struct {
	ID          uuid.UUID `db:"id"`
	Date        time.Time `db:"date"`
	Description string    `db:"description"`
	MoodLabel   MoodLabel `db:"mood_label"`
	User        *User     `db:"-"`
	Tags        []string  `db:"tags"`
	BaseModel
}

type Tag struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
	User *User     `db:"-"`
	BaseModel
}
type DaylogDTO struct {
	ID          uuid.UUID  `json:"id"`
	Date        *time.Time `json:"date"`
	Description *string    `json:"description,omitempty"`
	MoodLabel   *string    `json:"mood_label"`
	User        *UserDTO   `json:"user,omitempty"`
	Tags        []*string  `json:"tags"`
}

type TagDTO struct {
	ID   uuid.UUID `json:"id"`
	Name *string   `json:"name"`
	User *UserDTO  `json:"user,omitempty"`
}

func (t Tag) ToDTO() *TagDTO {
	return &TagDTO{
		ID:   t.ID,
		Name: &t.Name,
		User: t.User.ToDTO(),
	}
}

func (dto TagDTO) ToModel() *Tag {
	var model Tag

	if dto.Name != nil {
		model.Name = *dto.Name
	}

	if dto.User != nil {
		model.User = dto.User.ToModel()
	}

	return &model
}

func (d Daylog) ToDTO() *DaylogDTO {
	dto := DaylogDTO{}

	dto.ID = d.ID
	dto.Date = &d.Date
	dto.Description = &d.Description
	label := d.MoodLabel.String()
	dto.MoodLabel = &label
	dto.Tags = utils.StringSliceToPtrSlice(d.Tags)

	if d.User != nil {
		dto.User = d.User.ToDTO()
	}

	return &dto
}

func (dto DaylogDTO) ToModel() *Daylog {
	model := Daylog{}

	if dto.Date != nil {
		model.Date = *dto.Date
	}

	if dto.Description != nil {
		model.Description = *dto.Description
	}

	if dto.MoodLabel != nil {
		label := parseMoodLabel(*dto.MoodLabel)
		model.MoodLabel = label
	}

	if dto.User != nil {
		model.User = dto.User.ToModel()
	}

	if dto.Tags != nil {
		model.Tags = utils.PtrStringSliceToSlice(dto.Tags)
	}

	return &model
}

func (m MoodLabel) String() string {
	switch m {
	case MOOD_RUIM:
		return "RUIM"
	case MOOD_MEDIO:
		return "MEDIO"
	case MOOD_BOM:
		return "BOM"
	default:
		return "unknown"
	}
}

func parseMoodLabel(s string) MoodLabel {
	switch strings.ToLower(s) {
	case "RUIM":
		return MOOD_RUIM
	case "MEDIO":
		return MOOD_MEDIO
	case "BOM":
		return MOOD_BOM
	default:
		return 0
	}
}

func (d *Daylog) ValidateDaylog(v *validator.Validator) {
	v.Check(d.Date.IsZero() == false, "date", "must be provided")
	v.Check(d.MoodLabel >= MOOD_RUIM && d.MoodLabel <= MOOD_BOM,
		"mood_label", "invalid mood label")

	if d.Description != "" {
		v.Check(len(d.Description) <= 1000,
			"description", "must not be more than 1000 bytes long")
	}
}

func (model *Tag) ValidateTag(v *validator.Validator) {
	v.Check(model.Name != "", "name", "must be provided")
}
