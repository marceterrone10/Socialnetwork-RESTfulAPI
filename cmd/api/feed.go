package main

import (
	"net/http"
)

func (app *application) getFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posts, err := app.store.Posts.GetFeed(ctx, int64(1))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
