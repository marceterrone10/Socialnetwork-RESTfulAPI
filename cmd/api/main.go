package main

import (
	"log"

	"github.com/marceterrone10/social/internal/env"
	"github.com/marceterrone10/social/internal/store"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	store := store.NewStorage(nil) // instancia del store y creo un nuevo storage

	app := &application{
		config: cfg,
		store:  store, // paso el store a la aplicación
	}

	// mount the routes for the API
	mux := app.mount()
	// serve the API

	log.Fatal(app.serve(mux)) // log the error if the server fails to start
}
