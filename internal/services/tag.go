package services

import (
	"database/sql"
	"moodtracker/internal/models"
	"moodtracker/internal/models/filters"
	"moodtracker/internal/repositories"
	"moodtracker/utils/validator"

	"github.com/google/uuid"
)

type tagService struct {
	tag repositories.TagRepository
	db  *sql.DB
}

func NewTagService(
	tag repositories.TagRepository,
	db *sql.DB,
) *tagService {
	return &tagService{
		tag: tag,
		db:  db,
	}
}

type TagService interface {
	GetAllByDayLogID(
		dayLogID uuid.UUID,
		userID int64,
		f filters.Filters,
	) ([]*models.Tag, filters.Metadata, error)
	Save(model *models.Tag, userID uuid.UUID, v *validator.Validator) error
	FindByID(id, userID uuid.UUID) (*models.Tag, error)
	Update(model *models.Tag, userID uuid.UUID, v *validator.Validator) error
	Delete(id, userID uuid.UUID) error
	GetIDByNameOrCreate(
		v *validator.Validator,
		name string,
		userID uuid.UUID,
	) (uuid.UUID, error)
}
