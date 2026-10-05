package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"murtari.me/web/internal/photos"
)

type photoView struct {
	photos.Photo
	ThumbnailURL string
	FullSizeURL  string
}

type photosPage struct {
	Photos  []photoView
	View    string
	Current photoView
	Index   int
}

func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (app *application) loadManifest() error {
	mp := filepath.Join(app.photoDir, "manifest.json")
	data, err := os.ReadFile(mp)
	if err != nil {
		return err
	}

	var p []photos.Photo
	err = json.Unmarshal(data, &p)
	if err != nil {
		return err
	}

	app.manifest = p
	return nil
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, nil, "base", "pages/home.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func (app *application) apps(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, nil, "base", "pages/apps.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func (app *application) photos(w http.ResponseWriter, r *http.Request) {
	photos := app.manifest

	v := make([]photoView, len(photos))
	for i, p := range photos {
		v[i] = photoView{
			Photo:        p,
			ThumbnailURL: p.ThumbPath(),
			FullSizeURL:  p.FullPath(),
		}
	}

	p := photosPage{Photos: v, View: "gallery"}
	if r.URL.Query().Get("view") == "slideshow" && len(v) > 0 {
		idx, err := strconv.Atoi(r.URL.Query().Get("index"))
		if err != nil || idx < 0 || idx >= len(v) {
			idx = 0
		}

		p.Index = idx
		p.View = "slideshow"
		p.Current = v[idx]
	}

	t := "base"
	if isHTMXRequest(r) {
		switch r.Header.Get("HX-Target") {
		case "section#slideshow":
			t = "photos:slideshow"
		case "div#slideshow-subject":
			t = "photos:slideshow:swap-target"
		default:
			t = "photos:page"
		}
	}

	err := app.html.render(w, 200, p, t, "pages/photos.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
