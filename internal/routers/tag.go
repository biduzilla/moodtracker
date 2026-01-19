package routers

import (
	"moodtracker/internal/handlers"
	"moodtracker/internal/middleware"

	"github.com/go-chi/chi"
)

type tagRouter struct {
	tag handlers.TagHandler
	m   middleware.MiddlewareInterface
}

type TagRouter interface {
	TagRoutes(r chi.Router)
}

func NewTagRouter(
	tag handlers.TagHandler,
	m middleware.MiddlewareInterface,

) *tagRouter {
	return &tagRouter{
		tag: tag,
		m:   m,
	}
}

func (r *tagRouter) TagRoutes(router chi.Router) {
	router.Route("/tags", func(router chi.Router) {
		router.Use(r.m.RequireActivatedUser)

		router.Get("/{id}", r.tag.FindByID)
		router.Get("/user/{id}", r.tag.GetAllByUserID)
		router.Post("/", r.tag.Save)
		router.Put("/", r.tag.Update)
		router.Delete("/{id}", r.tag.Delete)
	})
}
