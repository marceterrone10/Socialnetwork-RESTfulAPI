package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Internal server error: %s, path: %s, error: %s", r.Method, r.URL.Path, err.Error())
	errorJSON(w, http.StatusInternalServerError, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Not found error: %s, path: %s, error: %s", r.Method, r.URL.Path, err.Error())
	errorJSON(w, http.StatusNotFound, err.Error())
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Bad request error: %s, path: %s, error: %s", r.Method, r.URL.Path, err.Error())
	errorJSON(w, http.StatusBadRequest, err.Error())
}
