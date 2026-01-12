package models

import (
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
	MoodScore   int       `db:"mood_score"`
	Description string    `db:"description"`
	MoodLabel   MoodLabel `db:"mood_label"`
	User        *User     `db:"-"`
	BaseModel
}

type Tag struct {
	ID     uuid.UUID `db:"id"`
	Tag    string    `db:"tag"`
	Daylog *Daylog   `db:"-"`
	BaseModel
}
type DaylogDTO struct {
	ID          uuid.UUID  `json:"id"`
	Date        *time.Time `json:"date"`
	MoodScore   *int       `json:"mood_score"`
	Description *string    `json:"description,omitempty"`
	MoodLabel   *string    `json:"mood_label"`
	User        *UserDTO   `json:"user,omitempty"`
}

type TagDTO struct {
	ID     uuid.UUID  `json:"id"`
	Tag    *string    `json:"tag"`
	Daylog *DaylogDTO `json:"day_log"`
}

func (t *Tag) ToDTO() *TagDTO {
	if t == nil {
		return nil
	}

	return &TagDTO{
		ID:     t.ID,
		Tag:    &t.Tag,
		Daylog: t.Daylog.ToDTO(),
	}
}

func (dto *TagDTO) ToModel() *Tag {
	if dto == nil {
		return nil
	}

	var model Tag

	if dto.Tag != nil {
		model.Tag = *dto.Tag
	}

	if dto.Daylog != nil {
		model.Daylog = dto.Daylog.ToModel()
	}

	return &model
}

func (d *Daylog) ToDTO() *DaylogDTO {
	dto := DaylogDTO{}

	dto.ID = d.ID
	dto.Date = &d.Date
	dto.MoodScore = &d.MoodScore
	dto.Description = &d.Description
	label := d.MoodLabel.String()
	dto.MoodLabel = &label

	if d.User != nil {
		dto.User = d.User.ToDTO()
	}

	return &dto
}

func (dto *DaylogDTO) ToModel() *Daylog {
	model := Daylog{}

	if dto.Date != nil {
		model.Date = *dto.Date
	}

	if dto.MoodScore != nil {
		model.MoodScore = *dto.MoodScore
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

	v.Check(d.MoodScore >= 1 && d.MoodScore <= 5,
		"mood_score", "must be between 1 and 5")

	v.Check(d.MoodLabel >= MOOD_RUIM && d.MoodLabel <= MOOD_BOM,
		"mood_label", "invalid mood label")

	if d.Description != "" {
		v.Check(len(d.Description) <= 1000,
			"description", "must not be more than 1000 bytes long")
	}
}

func (model *Tag) ValidateTag(v *validator.Validator) {
	v.Check(model.Tag != "", "tag", "must be provided")
}
