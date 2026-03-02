package errors

import (
	"fmt"
	"gateway/internal/jsonlog"
	"gateway/internal/utils"
	"net/http"
)

type errorHandler struct {
	logger jsonlog.Logger
}

func NewErrorHandler(logger jsonlog.Logger) *errorHandler {
	return &errorHandler{
		logger: logger,
	}
}

type ErrorHandlerInterface interface {
	RateLimitExceededResponse(w http.ResponseWriter, r *http.Request)
	ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error)
	ProxyError(w http.ResponseWriter, r *http.Request, err error)
	NotFoundResponse(w http.ResponseWriter, r *http.Request)
}

func (e *errorHandler) HandlerError(w http.ResponseWriter, r *http.Request, err error) {

	switch {
	default:
		e.ServerErrorResponse(w, r, err)
	}
}

func (e *errorHandler) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	e.errorHandler(w, r, http.StatusNotFound, message)
}

func (e *errorHandler) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceed"
	e.errorHandler(w, r, http.StatusTooManyRequests, message)
}

func (e *errorHandler) ProxyError(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)
	message := fmt.Sprintf("proxy error: %v", err)
	e.errorHandler(w, r, http.StatusBadGateway, message)
}

func (e *errorHandler) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	e.errorHandler(w, r, http.StatusInternalServerError, message)
}

func (e *errorHandler) errorHandler(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := utils.Envelope{"error": message}
	err := utils.WriteJSON(w, status, env, nil)
	if err != nil {
		e.logError(r, err)
		w.WriteHeader(500)
	}
}

func (e *errorHandler) logError(r *http.Request, err error) {
	e.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}
