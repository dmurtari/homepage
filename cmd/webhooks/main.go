package main

import (
	"log/slog"
	"net/http"
	"os"
)

type application struct {
	logger *slog.Logger
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/workout", app.workout)

	logger.Info("starting server", "port", 5052)
	err := http.ListenAndServe(":5052", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func (app *application) workout(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
