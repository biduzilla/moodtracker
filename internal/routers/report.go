package routers

import (
	"moodtracker/internal/handlers"
	"moodtracker/internal/middleware"

	"github.com/go-chi/chi"
)

type reportRouter struct {
	report handlers.ReportHandler
	m      middleware.MiddlewareInterface
}

type ReportRouter interface {
	ReportRoutes(r chi.Router)
}

func NewReportRouter(
	report handlers.ReportHandler,
	m middleware.MiddlewareInterface,

) *reportRouter {
	return &reportRouter{
		report: report,
		m:      m,
	}
}

func (r *reportRouter) ReportRoutes(router chi.Router) {
	router.Route("/reports", func(router chi.Router) {
		router.Use(r.m.RequireActivatedUser)

		router.Get("/monthly", r.report.GetMonthlyReport)
		router.Get("/tag", r.report.GetTagReport)
		router.Get("/mood", r.report.GetMoodReport)

	})
}
