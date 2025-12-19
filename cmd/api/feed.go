package main

import (
	"net/http"

	"github.com/marceterrone10/social/internal/store"
)

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
