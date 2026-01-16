package services

import (
	"database/sql"
	"moodtracker/internal/models"
	"moodtracker/internal/models/filters"
	"moodtracker/internal/repositories"
	"moodtracker/utils"
	e "moodtracker/utils/errors"
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
	GetAllByUserID(
		userID uuid.UUID,
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

func (s *tagService) GetAllByUserID(
	userID uuid.UUID,
	f filters.Filters,
) ([]*models.Tag, filters.Metadata, error) {
	return s.tag.GetAllByUserID(userID, f)
}

func (s *tagService) Save(model *models.Tag, userID uuid.UUID, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateTag(v); !v.Valid() {
			return e.ErrInvalidData
		}

		return s.tag.Insert(tx, model, userID)
	})
}
func (s *tagService) FindByID(id, userID uuid.UUID) (*models.Tag, error) {
	return s.tag.FindByID(id, userID)
}

func (s *tagService) Update(model *models.Tag, userID uuid.UUID, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateTag(v); !v.Valid() {
			return e.ErrInvalidData
		}

		return s.tag.Update(tx, model, userID)
	})
}

func (s *tagService) Delete(id, userID uuid.UUID) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		err := s.tag.Delete(tx, id, userID)
		if err != nil {
			return err
		}

		return s.tag.DeleteLogTagByTagID(tx, id)
	})
}

func (s *tagService) GetIDByNameOrCreate(
	v *validator.Validator,
	name string,
	userID uuid.UUID,
) (uuid.UUID, error) {
	v.Check(name != "", "name", "must be provided")
	if !v.Valid() {
		return uuid.Nil, e.ErrInvalidData
	}

	var tagID uuid.UUID

	err := utils.RunInTx(s.db, func(tx *sql.Tx) error {
		var err error
		tagID, err = s.tag.GetIDByNameOrCreate(tx, name, userID)
		return err
	})

	if err != nil {
		return uuid.Nil, err
	}

	return tagID, nil
}
