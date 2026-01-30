package services

import (
	"database/sql"
	"moodtracker/internal/config"
	"moodtracker/internal/jsonlog"
	"moodtracker/internal/models"
	"moodtracker/internal/repositories"
	"moodtracker/utils/validator"

	"github.com/google/uuid"
)

type GenericServiceInterface[
	T models.ModelInterface[D],
	D any,
] interface {
	Save(entity *T, userID uuid.UUID, v *validator.Validator) error
	FindByID(id, userID uuid.UUID) (*T, error)
	Update(entity *T, userID uuid.UUID, v *validator.Validator) error
	Delete(id, userID uuid.UUID) error
}

type Services struct {
	User   UserService
	Auth   AuthServiceInterface
	Daylog DaylogServices
	Tag    TagService
}

func NewServices(logger jsonlog.Logger, db *sql.DB, config config.Config) *Services {
	r := repositories.NewRepository(logger, db)
	userService := NewUserService(r.User, db)
	tagService := NewTagService(r.Tag, db)
	return &Services{
		User:   userService,
		Auth:   NewAuthService(userService, config),
		Daylog: NewDaylogService(r.DayLog, db, tagService),
		Tag:    tagService,
	}
}
