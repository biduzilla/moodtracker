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
		userID uuid.UUID,
		f filters.Filters,
	) ([]*models.Daylog, filters.Metadata, error)
	GetAllByYear(
		year int,
		userID uuid.UUID,
	) ([]*models.Daylog, error)
	GetByID(id, userID uuid.UUID) (*models.Daylog, error)
	InsertLogsTags(tx *sql.Tx, daylogID, tagID uuid.UUID) error
	InsertOrUpdate(tx *sql.Tx, model *models.Daylog, userID uuid.UUID) error
	Update(tx *sql.Tx, model *models.Daylog, userID uuid.UUID) error
	Delete(tx *sql.Tx, id uuid.UUID, userID uuid.UUID) error
	DeleteLogTagByDaylogID(tx *sql.Tx, daylogID uuid.UUID) error
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

/*
create unique index uniq_day_logs_user_date
on day_logs (user_id, date)
where deleted = false;

create unique index uniq_tags_name_user_not_deleted
on tags (lower(name), user_id)
where deleted = false;

*/

func (r *daylogRepository) GetAllByYear(
	year int,
	userID uuid.UUID,
) ([]*models.Daylog, error) {
	cols := strings.Join([]string{
		selectColumns(models.Daylog{}, "dl"),
	}, ", ")

	query := fmt.Sprintf(`
        SELECT
           	%s,
			ARRAY_AGG(t.name) as tags
        FROM day_logs dl
		LEFT JOIN log_tags lt ON dl.id = lt.log_id
		LEFT JOIN tags ut ON lt.tag_id = t.id
        WHERE
    		dl.date >= make_date(:year, 1, 1)
    		AND dl.date < make_date(:year + 1, 1, 1)
            AND dl.deleted = false
			and dl.user_id = :userID 
    `, cols)

	params := map[string]any{
		"year":   year,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return listQuery(
		r.db,
		query,
		args,
		func() *models.Daylog {
			return &models.Daylog{
				User: &models.User{},
			}
		},
	)
}

func (r *daylogRepository) GetAll(
	description, tag string,
	modelLabel *models.MoodLabel,
	date *time.Time,
	userID uuid.UUID,
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

func (r *daylogRepository) GetByID(id, userID uuid.UUID) (*models.Daylog, error) {
	cols := strings.Join([]string{
		selectColumns(models.Daylog{}, "dl"),
		selectColumns(models.User{}, "u"),
	}, ", ")

	query := fmt.Sprintf(`
	select 
	%s,
	ARRAY_AGG(t.name) as tags
	FROM day_logs dl
	LEFT JOIN log_tags lt ON dl.id = lt.log_id
	LEFT JOIN tags ut ON lt.tag_id = t.id
	WHERE
		dl.id = :id
		and dl.user_id = :userID
		AND dl.deleted = false
	GROUP BY %s
	`, cols, cols)

	params := map[string]any{
		"id":     id,
		"userID": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)
	return getByQuery[models.Daylog](r.db, query, args)
}

func (r *daylogRepository) InsertLogsTags(tx *sql.Tx, daylogID, tagID uuid.UUID) error {
	query := `
	insert into log_tags (
		log_id,
		tag_id
	)
	values (
		:logID,
		:tagID
	)
	`

	params := map[string]any{
		"logID": daylogID,
		"tagID": tagID,
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

func (r *daylogRepository) InsertOrUpdate(
	tx *sql.Tx,
	model *models.Daylog,
	userID uuid.UUID,
) error {

	query := `
	insert into day_logs (
		date,
		description,
		mood_label,
		user_id,
		created_by
	)
	values (
		:date,
		:description,
		:moodLabel,
		:userID,
		:userID
	)
	on conflict (user_id, date)
	do update set
		description = excluded.description,
		mood_label = excluded.mood_label,
		updated_at = now(),
		version = day_logs.version + 1
	returning
		id,
		created_at,
		version
	`

	start := sql.NullTime{
		Time:  model.Date,
		Valid: true,
	}

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

	return tx.QueryRowContext(ctx, query, args...).Scan(
		&model.ID,
		&model.CreatedAt,
		&model.Version,
	)
}

func (r *daylogRepository) Update(tx *sql.Tx, model *models.Daylog, userID uuid.UUID) error {
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

func (r *daylogRepository) Delete(tx *sql.Tx, id uuid.UUID, userID uuid.UUID) error {
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

func (r *daylogRepository) DeleteLogTagByDaylogID(tx *sql.Tx, daylogID uuid.UUID) error {
	query := `
	delete from log_tags
	where log_id = :id
	`
	params := map[string]any{
		"id": daylogID,
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
