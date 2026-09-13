package main

import (
	"log/slog"
	"net/http"
	"os"

	"murtari.me/web/assets"
)

type application struct {
	logger *slog.Logger
	html   *htmlRenderer
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	htmlRenderer, err := newHTMLRenderer(assets.HTMLFiles, "base.tmpl", "partials/*.tmpl")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		logger: logger,
		html:   htmlRenderer,
	}

	fileserver := http.FileServerFS(assets.StaticFiles)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /apps", app.apps)
	mux.HandleFunc("GET /photos", app.photos)

	logger.Info("starting server", "port", 5051)
	err = http.ListenAndServe(":5051", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
