package main

import (
	"time"

	"github.com/marceterrone10/social/internal/db"
	"github.com/marceterrone10/social/internal/env"
	"github.com/marceterrone10/social/internal/mailer"
	"github.com/marceterrone10/social/internal/store"
	"go.uber.org/zap"
)

const version = "1.0.0"

//	@title			Social API
//	@description	RESTful API for a social media platform example
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Type (Bearer) followed by the token
func main() {
	// seteo de la app
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("API_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:3000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5433/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxLifetime:  env.GetString("DB_MAX_LIFETIME", "1h"),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// instancia de la DB
	database, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxLifetime)
	if err != nil {
		logger.Panicln(err)
	}

	defer database.Close()
	logger.Info("Connected to the database")

	// instancia del store y creo un nuevo storage con la DB
	storage := store.NewStorage(database)

	// instancia del mailer
	mailer := mailer.NewSendGridMailer(cfg.mail.fromEmail, cfg.mail.sendGrid.apiKey)

	// instancia de la app
	app := &application{
		config: cfg,
		store:  storage, // paso el store a la aplicación
		logger: logger,
		mailer: mailer,
	}

	// mount the routes for the API
	mux := app.mount()

	logger.Fatal(app.serve(mux)) // log the error if the server fails to start
}
