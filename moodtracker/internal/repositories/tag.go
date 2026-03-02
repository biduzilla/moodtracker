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
	GetAllByUserID(
		userID uuid.UUID,
		f filters.Filters,
	) ([]*models.Tag, filters.Metadata, error)
	FindByID(tagID, userID uuid.UUID) (*models.Tag, error)
	Insert(
		tx *sql.Tx,
		model *models.Tag,
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
	DeleteLogTagByTagID(tx *sql.Tx, tagID uuid.UUID) error
	GetIDByNameOrCreate(
		tx *sql.Tx,
		name string,
		userID uuid.UUID,
	) (uuid.UUID, error)
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
		case "uniq_tags_name_user":
			return e.ValidationAlreadyExists("tag")
		}
		return err
	}

	return nil
}

func (r *tagRepository) GetAllByUserID(
	userID uuid.UUID,
	f filters.Filters,
) ([]*models.Tag, filters.Metadata, error) {
	cols := strings.Join([]string{
		selectColumns(models.Tag{}, "t"),
	}, ", ")

	query := fmt.Sprintf(`
        SELECT
            count(*) OVER(),
           	%s
        FROM tags t
        LEFT JOIN users u ON u.id = t.user_id
        WHERE
			t.user_id = :userID
            AND t.deleted = false
        ORDER BY
            t.%s %s,
            t.id ASC
        LIMIT :limit
        OFFSET :offset
    `, cols, f.SortColumn(), f.SortDirection())

	params := map[string]any{
		"userID": userID,
		"limit":  f.Limit(),
		"offset": f.Offset(),
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
				User: &models.User{},
			}
		},
	)
}

func (r *tagRepository) FindByID(tagID, userID uuid.UUID) (*models.Tag, error) {
	cols := strings.Join([]string{
		selectColumns(models.Tag{}, "t"),
		selectColumns(models.User{}, "u"),
	}, ", ")

	query := fmt.Sprintf(`
        SELECT
           	%s
        FROM tags t
        LEFT JOIN users u ON u.id = t.use_id
        WHERE
			t.user_id = :userID
			and t.id = :tagID
            AND t.deleted = false
    `, cols)

	params := map[string]any{
		"userID": userID,
		"tagID":  tagID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	return getByQuery[models.Tag](r.db, query, args)
}

func (r *tagRepository) GetIDByNameOrCreate(
	tx *sql.Tx,
	name string,
	userID uuid.UUID,
) (uuid.UUID, error) {

	query := `
	insert into tags(
		name,
		user_id,
		created_by
	)
	values (
		:name,
		:userId,
		:userId
	)
	on conflict (lower(name), user_id) 
		WHERE deleted = false
	do update set name = excluded.name
	returning id
	`
	params := map[string]any{
		"name":   name,
		"userId": userID,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	var id uuid.UUID

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *tagRepository) Insert(
	tx *sql.Tx,
	model *models.Tag,
	userID uuid.UUID,
) error {
	query := `
	INSERT INTO tags (
		name, 
		user_id, 
		created_by,
	)
	VALUES (
		:tag, 
		:userID, 
		:userID,
		)
	RETURNING id, created_at, version
	`
	params := map[string]any{
		"tag":    model.Name,
		"userID": userID,
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
	UPDATE tags 
	SET
		name = :name,
		updated_at = NOW(),
		updated_by = :userID,
		version = version + 1
	WHERE
		user_id = :userID
		AND id = :id
		AND version = :version
		AND deleted = false
	RETURNING version
	`

	params := map[string]any{
		"id":      model.ID,
		"name":    model.Name,
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
	UPDATE tags 
	SET
		deleted = true,
		updated_at = NOW(),
		updated_by = :userID
	WHERE
		user_id = :userID
		AND id = :id
		AND deleted = false
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

func (r *tagRepository) DeleteLogTagByTagID(tx *sql.Tx, tagID uuid.UUID) error {
	query := `
	delete from log_tags
	where tag_id = :id
	`
	params := map[string]any{
		"id": tagID,
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
