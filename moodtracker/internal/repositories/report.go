package repositories

import (
	"database/sql"
	"moodtracker/internal/jsonlog"
	"moodtracker/internal/models"
	"moodtracker/utils"
	"time"

	"github.com/google/uuid"
)

type reportRepository struct {
	db     *sql.DB
	logger jsonlog.Logger
}

type ReportRepository interface {
	GetMonthlyReport(
		year int,
		month int,
		userID uuid.UUID,
	) (*models.MonthlyReport, error)

	GetTagReport(
		tag string,
		userID uuid.UUID,
	) (*models.TagReport, error)

	GetMoodReport(
		moodLabel models.MoodLabel,
		userID uuid.UUID,
	) (*models.MoodReport, error)
}

func NewReportRepository(
	db *sql.DB,
	logger jsonlog.Logger,
) *reportRepository {
	return &reportRepository{
		db:     db,
		logger: logger,
	}
}

type monthlyMoodRow struct {
	Year       int
	Month      int
	MoodLabel  models.MoodLabel `db:"mood_label"`
	Count      int              `db:"count"`
	Percentage float64          `db:"percentage"`
}

type tagMoodRow struct {
	Tag        string           `db:"tag"`
	MoodLabel  models.MoodLabel `db:"mood_label"`
	Count      int              `db:"count"`
	Percentage float64          `db:"percentage"`
}

type moodTagRow struct {
	MoodLabel  models.MoodLabel `db:"mood_label"`
	Tag        string           `db:"tag"`
	Count      int              `db:"count"`
	Percentage float64          `db:"percentage"`
}

type monthlyTagRow struct {
	Tag   string `db:"tag"`
	Count int    `db:"count"`
}

func (r *reportRepository) GetMonthlyReport(
	year int,
	month int,
	userID uuid.UUID,
) (*models.MonthlyReport, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	moodQuery := `
	select 
		dl.mood_label,
		count(*) as count,
		round(
			count(*) *100.0 /
			sum(count(*)) over(),
			2
		) as percentage
	from day_logs dl
	where
		dl.user_id = :userID
		and dl.deleted = false
		and dl.date >= :startDate
		and dl.date < :endDate
	group by dl.mood_label
	order by dl.mood_label
	`

	params := map[string]any{
		"userID":    userID,
		"startDate": start,
		"endDate":   end,
	}

	moodQuery, moodArgs := namedQuery(moodQuery, params)
	r.logger.PrintInfo(utils.MinifySQL(moodQuery), nil)

	moodFactory := func() *monthlyMoodRow {
		return &monthlyMoodRow{}
	}

	moodList, err := listQuery(r.db, moodQuery, moodArgs, moodFactory)
	if err != nil {
		return nil, err
	}

	tagQuery := `
	select
		t.name as tag,
		count(*) as count
	from day_logs dl
	join log_tags lt on lt.log_id = dl.id
	join tags t on t.id = lt.tag_id
	where
		dl.user_id = :userID
		and dl.deleted = false
		and t.deleted = false
		and dl.date >= :startDate
		and dl.date < :endDate
	GROUP BY t.name
	ORDER BY count DESC;
	`

	tagQuery, tagArgs := namedQuery(tagQuery, params)
	r.logger.PrintInfo(utils.MinifySQL(tagQuery), nil)

	tagFactory := func() *monthlyTagRow {
		return &monthlyTagRow{}
	}

	tagList, err := listQuery(r.db, tagQuery, tagArgs, tagFactory)
	if err != nil {
		return nil, err
	}

	report := &models.MonthlyReport{
		Year:          year,
		Month:         month,
		Distribuition: []models.MoodDistribuition{},
		Tags:          []models.CountTags{},
	}

	for _, row := range moodList {
		report.Distribuition = append(report.Distribuition,
			models.MoodDistribuition{
				MoodLabel:  row.MoodLabel,
				Count:      row.Count,
				Percentage: row.Percentage,
			},
		)
	}

	for _, row := range tagList {
		report.Tags = append(report.Tags,
			models.CountTags{
				Tag:   row.Tag,
				Count: row.Count,
			},
		)
	}

	return report, nil
}

func (r *reportRepository) GetTagReport(
	tag string,
	userID uuid.UUID,
) (*models.TagReport, error) {
	query := `
	select
		t.name as tag,
		dl.mood_label,
		count(*)as count,
		round(
			count(*) * 100.0 /
			sum(count(*)) over(),
			2
		) as percentage
	from tags t
	join log_tags lt on lt.tag_id = t.id
	join day_logs dl on dl.id = lt.log_id
	where
		dl.user_id = :userID
		and dl.deleted = false
		and t.deleted = false
		and lower(t.name) = lower(:tag)
	group by t.name,dl.mood_label
	order by dl.mood_label
	`

	params := map[string]any{
		"userID": userID,
		"tag":    tag,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	tagMoodFactory := func() *tagMoodRow {
		return &tagMoodRow{}
	}

	tagList, err := listQuery(r.db, query, args, tagMoodFactory)
	if err != nil {
		return nil, err
	}

	tagReport := &models.TagReport{
		Tag:           tag,
		Distribuition: []models.MoodDistribuition{},
	}

	for _, row := range tagList {
		tagReport.Distribuition = append(tagReport.Distribuition,
			models.MoodDistribuition{
				MoodLabel:  row.MoodLabel,
				Count:      row.Count,
				Percentage: row.Percentage,
			},
		)
	}

	return tagReport, nil
}

func (r *reportRepository) GetMoodReport(
	moodLabel models.MoodLabel,
	userID uuid.UUID,
) (*models.MoodReport, error) {
	query := `
	select
		t.name as tag,
		count(*) as count,
		round(
			count(*) * 100.0 /
			sum(count(*)) over (),
			2
		)as percentage
	from day_logs dl
	join log_tags lt on lt.log_id = dl.id
	join tags on t.id = lt.tag_id
	where
		dl.user_id = :userID
		and dl.deleted = false
		and t.deleted = false
		and dl.mood_label = :mood
	group by t.name
	order by count desc
	`

	params := map[string]any{
		"userID": userID,
		"mood":   moodLabel,
	}

	query, args := namedQuery(query, params)
	r.logger.PrintInfo(utils.MinifySQL(query), nil)

	moodTagFactory := func() *moodTagRow {
		return &moodTagRow{}
	}

	list, err := listQuery(r.db, query, args, moodTagFactory)
	if err != nil {
		return nil, err
	}

	moodReport := &models.MoodReport{
		MoodLabel:     moodLabel,
		Distribuition: []models.TagDistribuition{},
	}

	for _, row := range list {
		moodReport.Distribuition = append(moodReport.Distribuition,
			models.TagDistribuition{
				Tag:        row.Tag,
				Count:      row.Count,
				Percentage: row.Percentage,
			},
		)
	}

	return moodReport, nil
}
