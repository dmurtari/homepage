package main

import (
	"log/slog"
	"net/http"
	"os"

	"murtari.me/web/assets"
)

type application struct {
	logger   *slog.Logger
	html     *htmlRenderer
	photoDir string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	htmlRenderer, err := newHTMLRenderer(assets.HTMLFiles, "base.tmpl", "partials/*.tmpl")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	photoDir := os.Getenv("PHOTO_DIR")
	if photoDir == "" {
		photoDir = "./tmp/out"
	}

	app := &application{
		logger:   logger,
		html:     htmlRenderer,
		photoDir: photoDir,
	}

	fileserver := http.FileServerFS(assets.StaticFiles)
	photoFileServer := http.FileServer(http.Dir(app.photoDir))

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))
	mux.Handle("GET /media/photos/", http.StripPrefix("/media/photos", photoFileServer))
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /apps", app.apps)
	mux.HandleFunc("GET /photos", app.photos)
	mux.HandleFunc("GET /photos/subject", app.photoSlideshowSubject)

	logger.Info("starting server", "port", 5051)
	err = http.ListenAndServe(":5051", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
