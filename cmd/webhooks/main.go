package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"murtari.me/web/internal/database"
)

type application struct {
	logger *slog.Logger
	db     *sql.DB
	apiKey string
	secret string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbPath := os.Getenv("ACTIVITIES_DB_PATH")
	if dbPath == "" {
		logger.Error("ACTIVITIES_DB_PATH is not set")
		os.Exit(1)
	}

	apiKey := os.Getenv("WEBHOOK_API_KEY")
	if apiKey == "" {
		logger.Error("WEBHOOK_API_KEY is not set")
		os.Exit(1)
	}

	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		logger.Error("WEBHOOK_SECRET is not set")
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
		apiKey: apiKey,
		secret: secret,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/activities", app.activities)

	logger.Info("starting server", "port", 5052)
	err = http.ListenAndServe(":5052", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
