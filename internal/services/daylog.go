package services

import (
	"database/sql"
	"moodtracker/internal/models"
	"moodtracker/internal/repositories"
	"moodtracker/utils"
	e "moodtracker/utils/errors"
	"moodtracker/utils/validator"

	"github.com/google/uuid"
)

type daylogServices struct {
	daylog repositories.DaylogRepository
	db     *sql.DB
}

func NewDaylogService(
	daylog repositories.DaylogRepository,
	db *sql.DB,
) *daylogServices {
	return &daylogServices{
		daylog: daylog,
		db:     db,
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

		return s.daylog.Insert(tx, model, userID)
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
		return s.daylog.Delete(tx, id, userID)
	})
}
