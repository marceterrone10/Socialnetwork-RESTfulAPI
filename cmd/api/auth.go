package main

import (
	"net/http"

	"github.com/marceterrone10/social/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// registerUserHandler godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user with a username, email, and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"Register user payload"
//	@Success		201		{object}	store.User			"User registered successfully"
//	@Failure		400		{object}	error				"Bad request"
//	@Failure		500		{object}	error				"Internal server error"
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	app.store.Users.CreateInvitation(ctx, user, "string123")

	if err := app.writeResponse(w, http.StatusCreated, "User registered successfully"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
