package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marceterrone10/social/internal/store"
)

type userKey string

const userCtx userKey = "user" // clave para el contexto del usuario

// GetUser godoc
//
//	@Summary		Get a user by ID
//	@Description	Get a user by ID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		500	{object}	error
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r.Context())

	if err := app.writeResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		Follow a user
//	@Description	Follow a user by ID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int			true	"User ID"
//	@Param			payload	body		FollowUser	true	"Follow User Payload"
//	@Success		200		{string}	string		"User followed successfully"
//	@Failure		400		{object}	error		"User payload missing"
//	@Failure		404		{object}	error		"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r.Context())
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	err = app.store.Follows.Follow(ctx, followerUser.ID, followedID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusOK, "User followed successfully"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UnfollowUser godoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int			true	"User ID"
//	@Param			body	body		FollowUser	true	"Unfollow user payload"
//	@Success		200		{string}	string		"User unfollowed successfully"
//	@Failure		400		{object}	error		"User payload missing"
//	@Failure		500		{object}	error		"Internal server error"
//	@Failure		404		{object}	error		"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r.Context())
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	err = app.store.Follows.Unfollow(ctx, followerUser.ID, followedID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusOK, "User unfollowed successfully"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extrae el id del usuario de la URL
		id, err := strconv.ParseInt((chi.URLParam(r, "id")), 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}
		ctx := r.Context()
		user, err := app.store.Users.GetById(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(ctx context.Context) *store.User {
	user, _ := ctx.Value(userCtx).(*store.User)
	return user
}
