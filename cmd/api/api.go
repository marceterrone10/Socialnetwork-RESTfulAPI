package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/marceterrone10/social/internal/store"
)

type application struct {
	config config
	store  store.Storage // inyeccion de dependencias, paso el store a la aplicación
}

type config struct {
	addr string
	db   dbConfig
	env  string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxLifetime  string
}

/* Mount the routes for the API */
func (app *application) mount() *chi.Mux {
	r := chi.NewRouter() // create a new router

	// use the middleware for the router
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Use(middleware.Timeout(60 * time.Second)) // middleware to timeout requests after 60 seconds

	// route the API to the healthcheck handler
	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)

			})
		})
		r.Route("/comments", func(r chi.Router) {
			r.Post("/", app.createCommentHandler)
		})
		r.Route("/users", func(r chi.Router) {
			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)

			})
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getFeedHandler)
			})
		})
	})
	return r
}

/* Serve the API */
func (app *application) serve(mux *chi.Mux) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Starting server on %s", app.config.addr)

	return srv.ListenAndServe()

}
