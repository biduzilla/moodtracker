package services

import (
	"database/sql"
	"fmt"
	"moodtracker/internal/models"
	"moodtracker/internal/repositories"
	"moodtracker/utils"
	e "moodtracker/utils/errors"
	"moodtracker/utils/validator"

	"github.com/google/uuid"
)

type daylogServices struct {
	daylog repositories.DaylogRepository
	tag    TagService
	db     *sql.DB
}

func NewDaylogService(
	daylog repositories.DaylogRepository,
	db *sql.DB,
	tag TagService,
) *daylogServices {
	return &daylogServices{
		daylog: daylog,
		db:     db,
		tag:    tag,
	}
}

type DaylogServices interface {
	GetAllByYear(
		year int,
		userID uuid.UUID,
	) ([]*models.Daylog, error)
	Save(model *models.Daylog, userID uuid.UUID, v *validator.Validator) error
	FindByID(id, userID uuid.UUID) (*models.Daylog, error)
	Update(model *models.Daylog, userID uuid.UUID, v *validator.Validator) error
	Delete(id, userID uuid.UUID) error
}

func (s *daylogServices) GetAllByYear(
	year int,
	userID uuid.UUID,
) ([]*models.Daylog, error) {
	return s.daylog.GetAllByYear(year, userID)
}

func (s *daylogServices) Save(model *models.Daylog, userID uuid.UUID, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateDaylog(v); !v.Valid() {
			return e.ErrInvalidData
		}

		err := s.daylog.InsertOrUpdate(tx, model, userID)
		if err != nil {
			return err
		}

		fmt.Printf("DEBUG: After InsertOrUpdate, ID: %v\n", model.ID)
		fmt.Printf("DEBUG: Tags count: %d\n", len(model.Tags))

		tagsIDs := make([]uuid.UUID, 0, len(model.Tags))

		for _, tagName := range model.Tags {
			fmt.Printf("DEBUG: Processing tag: %s\n", tagName)
			tagID, err := s.tag.GetIDByNameOrCreate(v, tagName, userID)
			if err != nil {
				return err
			}
			fmt.Printf("DEBUG: Got tag ID: %v\n", tagID)
			tagsIDs = append(tagsIDs, tagID)
		}

		for _, tagID := range tagsIDs {
			if err := s.daylog.InsertLogsTags(tx, model.ID, tagID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *daylogServices) FindByID(id, userID uuid.UUID) (*models.Daylog, error) {
	return s.daylog.GetByID(id, userID)
}

func (s *daylogServices) Update(model *models.Daylog, userID uuid.UUID, v *validator.Validator) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		if model.ValidateDaylog(v); !v.Valid() {
			return e.ErrInvalidData
		}

		return s.daylog.Update(tx, model, userID)
	})
}

func (s *daylogServices) Delete(id, userID uuid.UUID) error {
	return utils.RunInTx(s.db, func(tx *sql.Tx) error {
		err := s.daylog.Delete(tx, id, userID)
		if err != nil {
			return err
		}

		return s.daylog.DeleteLogTagByDaylogID(tx, id)
	})
}
