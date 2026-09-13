package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"murtari.me/web/internal/photos"
)

type photoView struct {
	photos.Photo
	ThumbnailURL string
	FullSizeURL  string
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
	mp := filepath.Join(app.photoDir, "manifest.json")
	data, err := os.ReadFile(mp)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	var photos []photos.Photo
	err = json.Unmarshal(data, &photos)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	v := make([]photoView, len(photos))
	for i, p := range photos {
		v[i] = photoView{
			Photo:        p,
			ThumbnailURL: fmt.Sprintf("/media/photos/%s/thumb.jpg", p.Slug),
			FullSizeURL:  fmt.Sprintf("/media/photos/%s/full.jpg", p.Slug),
		}
	}

	err = app.html.render(w, 200, v, "base", "pages/photos.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
