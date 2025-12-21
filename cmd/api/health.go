package main

import (
	"net/http"
)

// Healthcheck godoc
//
//	@Summary		Healthcheck the API
//	@Description	Check if the API is running and get system information
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	map[string]string	"Returns status, environment, and version"
//	@Failure		500	{object}	error				"Internal server error"
//	@Router			/healthcheck [get]
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
