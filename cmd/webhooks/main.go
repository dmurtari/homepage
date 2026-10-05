package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"murtari.me/web/internal/workouts"
)

type application struct {
	logger *slog.Logger
	db     *sql.DB
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbPath := os.Getenv("WORKOUT_DB_PATH")
	if dbPath == "" {
		os.Exit(1)
	}

	db, err := database.Open(dbPath, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	app := &application{
		logger: logger,
		db:     db,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/workout", app.workout)

	logger.Info("starting server", "port", 5052)
	err = http.ListenAndServe(":5052", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func (app *application) workout(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
