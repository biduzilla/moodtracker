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
	"github.com/lib/pq"
)

type tagRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

type TagRepository interface {
	GetAllByDayLogID(
		dayLogID,
		userID uuid.UUID,
		f filters.Filters,
	) ([]*models.Tag, filters.Metadata, error)
	Insert(
		tx *sql.Tx,
		model *models.Tag,
		dayLogID,
		userID uuid.UUID,
	) error
	Update(
		tx *sql.Tx,
		model *models.Tag,
		userID uuid.UUID,
	) error
	Delete(
		tx *sql.Tx,
		id,
		userID uuid.UUID,
	) error
}

func NewTagRepository(
	db *sql.DB,
	logger jsonlog.Logger,
) *tagRepository {
	return &tagRepository{
		db:     db,
		logger: logger,
	}
}

func parseTagConstraintError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Constraint {
		case "tag_description_key":
			return e.ValidationAlreadyExists("tag")
		}
		return err
	}

	return nil
}

func (r *tagRepository) GetAllByDayLogID(
	dayLogID,
	userID uuid.UUID,
	f filters.Filters,
) ([]*models.Tag, filters.Metadata, error) {
	cols := strings.Join([]string{
		selectColumns(models.Tag{}, "t"),
		selectColumns(models.Daylog{}, "dl"),
	}, ", ")

	query := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
           	%s
        FROM tags t
        LEFT JOIN day_logs dl ON dl.id = t.day_log_id
        WHERE
			t.day_log_id = :dayLogID
            AND t.deleted = false
			and dl.user_id = :userID
        ORDER BY
            t.%s %s,
            t.id ASC
        LIMIT :limit
        OFFSET :offset
    `, cols, f.SortColumn(), f.SortDirection())

	params := map[string]any{
		"dayLogID": dayLogID,
		"userID":   userID,
		"limit":    f.Limit(),
		"offset":   f.Offset(),
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return paginatedQuery(
		r.db,
		query,
		args,
		f,
		func() *models.Tag {
			return &models.Tag{
				Daylog: &models.Daylog{
					User: &models.User{},
				},
			}
		},
	)
}

func (r *tagRepository) Insert(
	tx *sql.Tx,
	model *models.Tag,
	dayLogID,
	userID uuid.UUID,
) error {
	query := `
	INSERT INTO tags (
		tag, 
		day_log_id, 
		user_id, 
		created_by,
	)
	VALUES (
		:tag, 
		:dayLogID, 
		:userID, 
		:userID,
		)
	RETURNING id, created_at, version
	`
	params := map[string]any{
		"tag":      model.Tag,
		"dayLogID": dayLogID,
		"userID":   userID,
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
		return parseTagConstraintError(err)
	}

	return nil
}

func (r *tagRepository) Update(
	tx *sql.Tx,
	model *models.Tag,
	userID uuid.UUID,
) error {
	query := `
	UPDATE tags t
	SET
		tag = :tag,
		updated_at = NOW(),
		updated_by = :userID,
		version = version + 1
	FROM day_logs dl
	WHERE
		t.day_log_id = dl.id
		AND dl.user_id = :userID
		AND t.id = :id
		AND t.version = :version
		AND t.deleted = false
	RETURNING t.version
	`

	params := map[string]any{
		"id":      model.ID,
		"tag":     model.Tag,
		"userID":  userID,
		"version": model.Version,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(&model.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrEditConflict
		}
		return parseTagConstraintError(err)
	}

	return nil
}

func (r *tagRepository) Delete(
	tx *sql.Tx,
	id,
	userID uuid.UUID,
) error {
	query := `
	UPDATE tags t
	SET
		deleted = true,
		updated_at = NOW(),
		updated_by = :userID
	FROM day_logs dl
	WHERE
		t.day_log_id = dl.id
		AND dl.user_id = :userID
		AND t.id = :id
		AND t.deleted = false
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
