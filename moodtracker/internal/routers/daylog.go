package routers

import (
	"moodtracker/internal/handlers"
	"moodtracker/internal/middleware"

	"github.com/go-chi/chi"
)

type daylogRouter struct {
	daylog handlers.DaylogHandler
	m      middleware.MiddlewareInterface
}

type DaylogRouter interface {
	DaylogRoutes(r chi.Router)
}

func NewDaylogRouter(
	daylog handlers.DaylogHandler,
	m middleware.MiddlewareInterface,

) *daylogRouter {
	return &daylogRouter{
		daylog: daylog,
		m:      m,
	}
}

func (r *daylogRouter) DaylogRoutes(router chi.Router) {
	router.Route("/day_logs", func(router chi.Router) {
		router.Use(r.m.RequireActivatedUser)

		router.Get("/{id}", r.daylog.FindByID)
		router.Get("/year", r.daylog.GetAllByYear)
		router.Post("/", r.daylog.Save)
		router.Put("/", r.daylog.Update)
		router.Delete("/{id}", r.daylog.Delete)
	})
}
