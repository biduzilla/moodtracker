package api

import (
	"context"
	"errors"
	"fmt"
	"gateway/internal/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func (app *application) Server() error {
	moodHandler := http.StripPrefix("/mood", app.moodProxy)

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/mood/"):
			moodHandler.ServeHTTP(w, r)
		case r.URL.Path == "/debug/vars":
			http.DefaultServeMux.ServeHTTP(w, r)
		default:
			app.errHandler.NotFoundResponse(w, r)
		}
	})

	handlerWithMiddleware := middleware.Chain(
		baseHandler,
		app.middleware.Logging,
		app.middleware.RateLimit,
		app.middleware.Metrics,
		app.middleware.RecoverPanic,
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.config.Server.Host, app.config.Server.Port),
		Handler:      handlerWithMiddleware,
		IdleTimeout:  time.Minute,
		ErrorLog:     log.New(app.Logger, "", 0),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Logger.PrintInfo("shutting down gateway", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.Logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.Logger.PrintInfo("starting gateway server", map[string]string{
		"addr":     srv.Addr,
		"proxy_to": app.config.Services.MoodTracker.URL,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.PrintInfo("stopped gateway server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
