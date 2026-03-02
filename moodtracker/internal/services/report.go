package services

import (
	"moodtracker/internal/models"
	"moodtracker/internal/repositories"

	"github.com/google/uuid"
)

type reportService struct {
	report repositories.ReportRepository
}

type ReportService interface {
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

func NewReportService(report repositories.ReportRepository) *reportService {
	return &reportService{
		report: report,
	}
}

func (s *reportService) GetMonthlyReport(
	year int,
	month int,
	userID uuid.UUID,
) (*models.MonthlyReport, error) {
	return s.report.GetMonthlyReport(year, month, userID)
}

func (s *reportService) GetTagReport(
	tag string,
	userID uuid.UUID,
) (*models.TagReport, error) {
	return s.report.GetTagReport(tag, userID)
}

func (s *reportService) GetMoodReport(
	moodLabel models.MoodLabel,
	userID uuid.UUID,
) (*models.MoodReport, error) {
	return s.report.GetMoodReport(moodLabel, userID)
}
