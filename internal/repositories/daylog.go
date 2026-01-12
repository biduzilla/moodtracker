package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"moodtracker/internal/jsonlog"
	"moodtracker/internal/models"
	"moodtracker/internal/models/filters"
	"moodtracker/utils"
	e "moodtracker/utils/errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type daylogRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

type DaylogRepository interface {
	GetAll(
		description, tag string,
		modelLabel *models.MoodLabel,
		date *time.Time,
		userID int64,
		f filters.Filters,
	) ([]*models.Daylog, filters.Metadata, error)
	GetByID(id uuid.UUID, userID int64) (*models.Daylog, error)
	Insert(tx *sql.Tx, model *models.Daylog, userID int64) error
	Update(tx *sql.Tx, model *models.Daylog, userID int64) error
	Delete(tx *sql.Tx, id, userID int64) error
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

func (r *daylogRepository) GetAll(
	description, tag string,
	modelLabel *models.MoodLabel,
	date *time.Time,
	userID int64,
	f filters.Filters,
) ([]*models.Daylog, filters.Metadata, error) {
	cols := strings.Join([]string{
		selectColumns(models.Daylog{}, "dl"),
		selectColumns(models.User{}, "u"),
	}, ", ")

	query := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
           	%s
        FROM day_logs dl
		left join tags t on t.day_log_id = dl.id
        WHERE
            (to_tsvector('simple', dl.description) @@ plainto_tsquery('simple', :description) OR :description = '')
			and (:modelLabel is nil or dl.model_label = :modelLabel)
			AND (:date::timestamptz IS NULL OR dl.date >= :date::timestamptz)
            AND dl.deleted = false
			and dl.user_id = :userID
        ORDER BY
            dl.%s %s,
            dl.id ASC
        LIMIT :limit
        OFFSET :offset
    `, cols, f.SortColumn(), f.SortDirection())

	start := sql.NullTime{}
	if date != nil {
		start.Valid = true
		start.Time = *date
	}

	params := map[string]any{
		"description": description,
		"modelLabel":  modelLabel,
		"date":        start,
		"userID":      userID,
		"limit":       f.Limit(),
		"offset":      f.Offset(),
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return paginatedQuery(
		r.db,
		query,
		args,
		f,
		func() *models.Daylog {
			return &models.Daylog{
				User: &models.User{},
			}
		},
	)
}

func (r *daylogRepository) GetByID(id uuid.UUID, userID int64) (*models.Daylog, error) {
	cols := strings.Join([]string{
		selectColumns(models.Daylog{}, "dl"),
		selectColumns(models.User{}, "u"),
	}, ", ")

	query := fmt.Sprintf(`
	select 
	%s 
	FROM day_logs dl
	WHERE
		dl.id = :id
		and dl.user_id = :userID
		AND dl.deleted = false
	`, cols)

	params := map[string]any{
		"id":     id,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	return getByQuery[models.Daylog](r.db, query, args)
}

func (r *daylogRepository) Insert(tx *sql.Tx, model *models.Daylog, userID int64) error {
	query := `
	INSERT INTO day_logs (
		date, 
		description, 
		mood_label, 
		user_id, 
		created_by,
	)
	VALUES (
		:date, 
		:description, 
		:moodLabel, 
		:userID, 
		:userID,
		)
	RETURNING id, created_at, version
	`
	start := sql.NullTime{}
	start.Valid = true
	start.Time = model.Date

	params := map[string]any{
		"date":        start,
		"description": model.Description,
		"moodLabel":   model.MoodLabel,
		"userID":      userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&model.ID,
		&model.CreatedAt,
		&model.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}
		return err
	}

	return nil
}

func (r *daylogRepository) Update(tx *sql.Tx, model *models.Daylog, userID int64) error {
	query := `
	UPDATE day_logs SET
		date = :date,
		description = :date,
		mood_label = :mood_label,
		updated_at = NOW(),
		updated_by = :userID,
		version = version + 1
	WHERE
		id = :id
		AND version = :version
	RETURNING version`

	start := sql.NullTime{}
	start.Valid = true
	start.Time = model.Date

	params := map[string]any{
		"date":        start,
		"description": model.Description,
		"moodLabel":   model.MoodLabel,
		"userID":      userID,
		"version":     model.BaseModel.Version,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&model.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}
		return err
	}
	return nil
}

func (r *daylogRepository) Delete(tx *sql.Tx, id, userID int64) error {
	query := `
	UPDATE day_logs set
		deleted = true
	where 
		id = :id
		and user_id = :userID
	`

	params := map[string]any{
		"id":     id,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return e.ErrRecordNotFound
	}

	return nil
}
