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
	Tags        []*Tag    `db:"-"`
	BaseModel
}

type Tag struct {
	ID  uuid.UUID `db:"id"`
	Tag string    `db:"tag"`
	BaseModel
}
type DaylogDTO struct {
	ID          uuid.UUID `json:"id"`
	Date        time.Time `json:"date"`
	MoodScore   int       `json:"mood_score"`
	Description string    `json:"description,omitempty"`
	MoodLabel   string    `json:"mood_label"`
	User        *UserDTO  `json:"user,omitempty"`
	Tags        []*TagDTO `json:"tags"`
}

type TagDTO struct {
	ID  uuid.UUID `json:"id"`
	Tag string    `json:"tag"`
}

func (t *Tag) ToDTO() *TagDTO {
	if t == nil {
		return nil
	}

	return &TagDTO{
		ID:  t.ID,
		Tag: t.Tag,
	}
}

func (dto *TagDTO) ToModel() *Tag {
	if dto == nil {
		return nil
	}

	return &Tag{
		ID:  dto.ID,
		Tag: dto.Tag,
	}
}

func (d *Daylog) ToDTO() *DaylogDTO {
	if d == nil {
		return nil
	}

	dto := &DaylogDTO{
		ID:          d.ID,
		Date:        d.Date,
		MoodScore:   d.MoodScore,
		Description: d.Description,
		MoodLabel:   d.MoodLabel.String(),
	}

	if d.User != nil {
		dto.User = d.User.ToDTO()
	}

	if len(d.Tags) > 0 {
		dto.Tags = make([]*TagDTO, 0, len(d.Tags))
		for _, tag := range d.Tags {
			dto.Tags = append(dto.Tags, tag.ToDTO())
		}
	}

	return dto
}

func (dto *DaylogDTO) ToModel() *Daylog {
	if dto == nil {
		return nil
	}

	model := &Daylog{
		ID:          dto.ID,
		Date:        dto.Date,
		MoodScore:   dto.MoodScore,
		Description: dto.Description,
		MoodLabel:   parseMoodLabel(dto.MoodLabel),
	}

	if dto.User != nil {
		model.User = dto.User.ToModel()
	}

	if len(dto.Tags) > 0 {
		model.Tags = make([]*Tag, 0, len(dto.Tags))
		for _, tag := range dto.Tags {
			model.Tags = append(model.Tags, tag.ToModel())
		}
	}

	return model
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

	if len(d.Tags) > 0 {
		for _, tag := range d.Tags {
			v.Check(tag.Tag != "", "tags", "tag name must be provided")
			v.Check(len(tag.Tag) <= 100,
				"tags", "tag must not be more than 100 bytes")
		}
	}
}
