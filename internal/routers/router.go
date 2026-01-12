package routers

import (
	"database/sql"
	"expvar"
	"moodtracker/internal/config"
	"moodtracker/internal/handlers"
	"moodtracker/internal/jsonlog"
	"moodtracker/internal/middleware"
	"moodtracker/utils/errors"
	"net/http"

	"github.com/go-chi/chi"
)

type Router struct {
	errResp errors.ErrorHandlerInterface
	m       middleware.MiddlewareInterface
	user    UserRoutesInterface
	auth    AuthRoutesInterface
}

func NewRouter(
	db *sql.DB,
	logger jsonlog.Logger,
	config config.Config,
) *Router {
	e := errors.NewErrorHandler(logger)
	h := handlers.NewHandler(db, e, config, logger)
	m := middleware.New(
		e,
		h.Service.User,
		h.Service.Auth,
		config,
	)
	return &Router{
		errResp: e,
		m:       m,
		user:    NewUserRouter(h.User),
		auth:    NewAuthRouter(h.Auth),
	}
}

func (router *Router) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(router.m.RecoverPanic)
	r.Use(router.m.Metrics)
	r.Use(router.m.RateLimit)
	r.Use(router.m.EnableCORS)
	r.Use(router.m.Authenticate)

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		router.errResp.NotFoundResponse(w, req)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		router.errResp.MethodNotAllowedResponse(w, req)
	})

	r.Route("/v1", func(r chi.Router) {
		r.Mount("/debug/vars", expvar.Handler())
		router.user.UserRoutes(r)
		router.auth.AuthRoutes(r)
	})

	return r
}
