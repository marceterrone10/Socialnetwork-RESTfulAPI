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
