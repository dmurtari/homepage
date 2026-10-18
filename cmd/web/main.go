package main

import (
	"log/slog"
	"net/http"
	"os"

	"murtari.me/web/assets"
	"murtari.me/web/internal/photos"
)

type application struct {
	logger   *slog.Logger
	html     *htmlRenderer
	photoDir string
	manifest []photos.Photo
}

func staticHandler() http.Handler {
	fs := http.FileServerFS(assets.StaticFiles)
	return http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=31536000, public, immutable")
		fs.ServeHTTP(w, r)
	}))
}

func robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFileFS(w, r, assets.StaticFiles, "robots.txt")
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

	if v := os.Getenv("MEDIA_BASE_URL"); v != "" {
		photos.MediaBaseURL = v
	}

	app := &application{
		logger:   logger,
		html:     htmlRenderer,
		photoDir: photoDir,
	}

	err = app.loadManifest()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", staticHandler())
	mux.HandleFunc("GET /robots.txt", robots)
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
