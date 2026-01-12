package services

import (
	"database/sql"
	"moodtracker/internal/config"
	"moodtracker/internal/jsonlog"
	"moodtracker/internal/models"
	"moodtracker/internal/repositories"
	"moodtracker/utils/validator"
)

type GenericServiceInterface[
	T models.ModelInterface[D],
	D any,
] interface {
	Save(entity *T, userID int64, v *validator.Validator) error
	FindByID(id, userID int64) (*T, error)
	Update(entity *T, userID int64, v *validator.Validator) error
	Delete(id, userID int64) error
}

type Services struct {
	User UserService
	Auth AuthServiceInterface
}

func NewServices(logger jsonlog.Logger, db *sql.DB, config config.Config) *Services {
	r := repositories.NewRepository(logger, db)
	userService := NewUserService(r.User, db)

	return &Services{
		User: userService,
		Auth: NewAuthService(userService, config),
	}
}
