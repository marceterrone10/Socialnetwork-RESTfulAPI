package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
	}
}
