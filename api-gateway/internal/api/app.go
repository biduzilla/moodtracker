package api

import (
	"gateway/internal/config"
	"gateway/internal/jsonlog"
	"gateway/internal/proxy"
	"gateway/internal/utils/errors"
	"net/http"
	"os"
	"sync"
)

type application struct {
	config    config.Config
	Logger    jsonlog.Logger
	wg        sync.WaitGroup
	moodProxy http.Handler
}

func NewApp() *application {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	errHandler := errors.NewErrorHandler(logger)
	cfg, err := config.Load()

	if err != nil {
		logger.PrintError(err, nil)
		return nil
	}

	moodProxy, err := proxy.New(proxy.Service{
		URL:        cfg.Services.MoodTracker.URL,
		Timeout:    cfg.Services.MoodTracker.Timeout,
		ErrHandler: errHandler,
	})

	if err != nil {
		logger.PrintError(err, nil)
		return nil
	}

	logger.PrintInfo("gateway initialized", map[string]string{
		"proxy_to": cfg.Services.MoodTracker.URL,
	})

	return &application{
		config:    *cfg,
		Logger:    logger,
		moodProxy: moodProxy,
	}
}
