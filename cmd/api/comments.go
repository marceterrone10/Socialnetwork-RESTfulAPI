package main

import (
	"net/http"

	"github.com/marceterrone10/social/internal/store"
)

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
	PostID  int64  `json:"post_id" validate:"required,min=1"`
	UserID  int64  `json:"user_id" validate:"required,min=1"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	comment := &store.Comment{
		Content: payload.Content,
		PostID:  payload.PostID,
		UserID:  payload.UserID,
	}

	ctx := r.Context()

	if err := app.store.Comments.Create(ctx, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
