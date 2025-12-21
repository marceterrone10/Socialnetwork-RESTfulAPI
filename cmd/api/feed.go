package main

import (
	"net/http"

	"github.com/marceterrone10/social/internal/store"
)

// GetFeed godoc
//
//	@Summary		Get a feed of posts
//	@Description	Get a feed of posts for the current user
//	@Tags			Feed
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int			false	"Limit the number of posts returned"
//	@Param			offset	query		int			false	"Offset the number of posts returned"
//	@Param			sort	query		string		false	"Sort the posts by created_at in ascending or descending order"
//	@Success		200		{array}		store.Post	"Feed of posts"
//	@Failure		400		{object}	error		"Bad request"
//	@Failure		500		{object}	error		"Internal server error"
//	@Router			/users/feed [get]
func (app *application) getFeedHandler(w http.ResponseWriter, r *http.Request) {

	fq := store.PaginatedQuery{ // default values
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.ParseURLParams(r) // parseo los parametros de la url y los reemplazo en fq
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	posts, err := app.store.Posts.GetFeed(ctx, int64(1), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
