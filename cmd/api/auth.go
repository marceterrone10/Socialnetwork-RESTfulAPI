package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/marceterrone10/social/internal/mailer"
	"github.com/marceterrone10/social/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

type CreateUserTokenPayload struct {
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
//	@Success		201		{object}	UserWithToken		"User registered successfully"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
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

	plainToken := uuid.New().String()

	// store
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// store the user
	err := app.store.Users.CreateInvitation(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDuplicateEmail):
			app.badRequestError(w, r, err)
		case errors.Is(err, store.ErrDuplicateUsername):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	// enviar email
	err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending email", "error", err)

		//rollback de la creación del usuario si el email no se envía (patron SAGA)
		// Es un patron de secuencia de transacciones locales. Cada una actualiza la DB local mediante transacciones ACID y ejecuta un evento para activar la siguiente. Si una falla, la saga ejecuta una serie de transacciones compensatorias que deshacen los cambios realizados por las transacciones anteriores.
		// Si falla el envio de email, hay que borrar todo lo que haya en las tablas de la DB de user_invitations con respecto al usuario y tambien borrar al propio usuario de la tabla users.
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", err)

		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := app.store.Users.ActivateUser(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.writeResponse(w, http.StatusNoContent, "User activated"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// createTokenHandler godoc
//
//	@Summary		Create a new token
//	@Description	Create a new token for a user
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"Create token payload"
//	@Success		200		{string}	string					"Token created successfully"
//	@Failure		400		{string}	error					"Bad request"
//	@Failure		500		{string}	error					"Internal server error"
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parsear el payload
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	// fetch de la DB al user (chequear si existe)
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorizedError(w, r, err)
		return
	}

	// generar el token -> jwt
	claims := jwt.MapClaims{
		"sub": user.ID,
		"aud": app.config.auth.token.aud,
		"iss": app.config.auth.token.iss,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// enviarlo al cliente
	if err := app.writeResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
