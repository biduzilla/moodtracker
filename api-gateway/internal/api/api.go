package api

import (
	"fmt"
	"gateway/internal/config"
	"gateway/internal/jsonlog"
	"gateway/internal/proxy"
	"gateway/internal/utils/errors"
	"net/http"
	"os"
)

func New() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	errHandler := errors.NewErrorHandler(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.PrintError(err, nil)
		return
	}

	moodProxy, err := proxy.New(proxy.Service{
		URL:        cfg.Services.MoodTracker.URL,
		Timeout:    cfg.Services.MoodTracker.Timeout,
		ErrHandler: errHandler,
	})

	if err != nil {
		logger.PrintError(err, nil)
	}

	router := http.NewServeMux()
	router.Handle("/mood/", moodProxy)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	server.ListenAndServe()
}
