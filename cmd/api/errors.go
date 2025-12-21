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
