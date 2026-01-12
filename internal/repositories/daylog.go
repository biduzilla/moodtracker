package repositories

import (
	"database/sql"
	"moodtracker/internal/jsonlog"
	"moodtracker/internal/models"

	"github.com/google/uuid"
)

type daylogRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

func NewDaylogRepository(
	db *sql.DB,
	logger jsonlog.Logger,
) *daylogRepository {
	return &daylogRepository{
		db:     db,
		logger: logger,
	}
}

func (r *daylogRepository) GetByID(ID, userID uuid.UUID) (*models.Daylog, error) {

}
