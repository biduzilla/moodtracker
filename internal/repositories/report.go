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

type monthlyMoodRow struct {
	Year       int
	Month      int
	MoodLabel  models.MoodLabel
	Count      int
	Percentage float64
}

type tagMoodRow struct {
	Tag        string  `db:"tag"`
	MoodLabel  int     `db:"mood_label"`
	Count      int     `db:"count"`
	Percentage float64 `db:"percentage"`
}

type moodTagRow struct {
	MoodLabel  int     `db:"mood_label"`
	Tag        string  `db:"tag"`
	Count      int     `db:"count"`
	Percentage float64 `db:"percentage"`
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
			sum(count(*) over()),
			2
		) as percentage
	from day_logs dl
	where
		dl:user_id = :userID
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
	r.logger.PrintInfo(utils.MinifySQL(moodQuery), nil)

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
