package models

type MonthlyReport struct {
	Year          int                 `db:"year"`
	Month         int                 `db:"month"`
	Distribuition []MoodDistribuition `db:"-"`
	Tags          []CountTags         `db:"-"`
}

type TagReport struct {
	Tag           string              `db:"tag"`
	Distribuition []MoodDistribuition `db:"-"`
}

type MoodReport struct {
	MoodLabel     `db:"mood_label"`
	Distribuition []TagDistribuition `db:"-"`
}

type MoodDistribuition struct {
	MoodLabel  `db:"mood_label"`
	Count      int     `db:"count"`
	Percentage float64 `db:"percentage"`
}

type TagDistribuition struct {
	Tag        string  `db:"tag"`
	Count      int     `db:"count"`
	Percentage float64 `db:"percentage"`
}

type CountTags struct {
	Tag   string `db:"tag"`
	Count int    `db:"count"`
}
