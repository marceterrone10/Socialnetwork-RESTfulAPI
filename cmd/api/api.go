package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/marceterrone10/social/docs"
	"github.com/marceterrone10/social/internal/mailer"
	"github.com/marceterrone10/social/internal/store"
	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware
	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage // inyeccion de dependencias, paso el store a la aplicación
	logger *zap.SugaredLogger
	mailer mailer.Client
}

type config struct {
	addr   string
	db     dbConfig
	env    string
	apiURL string
	mail   mailConfig
}

type mailConfig struct {
	exp       time.Duration
	sendGrid  sendGridConfig
	fromEmail string
}

type sendGridConfig struct {
	apiKey string
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

		docsURL := fmt.Sprintf("http://%s/v1/swagger/doc.json", app.config.apiURL)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsURL), //The url pointing to API definition
		))

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
			r.Put("/activate/{token}", app.activateUserHandler)
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

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
		})
	})
	return r
}

/* Serve the API */
func (app *application) serve(mux *chi.Mux) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("Starting server", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()

}
