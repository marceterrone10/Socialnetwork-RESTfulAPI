package main

import "log"

func main() {
	cfg := config{
		addr: ":8080",
	}

	app := &application{
		config: cfg,
	}

	// mount the routes for the API
	mux := app.mount()
	// serve the API

	log.Fatal(app.serve(mux)) // log the error if the server fails to start
}
