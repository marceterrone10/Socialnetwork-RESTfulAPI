package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Internal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	errorJSON(w, http.StatusInternalServerError, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	errorJSON(w, http.StatusNotFound, err.Error())
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	errorJSON(w, http.StatusBadRequest, err.Error())
}

func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	errorJSON(w, http.StatusUnauthorized, err.Error())
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Unauthorized basic auth error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	errorJSON(w, http.StatusUnauthorized, err.Error())
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Errorw("Forbidden error", "method", r.Method, "path", r.URL.Path)
	errorJSON(w, http.StatusForbidden, "Forbidden")
}
