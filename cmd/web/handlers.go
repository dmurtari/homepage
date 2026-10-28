package main

import (
	"context"
	"net/http"
	"strconv"

	"murtari.me/web/internal/photos"
)

type photoView struct {
	photos.Photo
	ThumbnailURL string
	FullSizeURL  string
}

type albumView struct {
	photos.Album
	photos []photoView
}

type photosPage struct {
	Photos  []photoView
	Albums  []albumView
	View    string
	Current photoView
	Index   int
}

func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (app *application) loadManifest() error {
	p, err := app.photoRepo.ListAll(context.TODO())
	if err != nil {
		return err
	}

	a, err := app.photoRepo.ListAlbums(context.TODO())
	if err != nil {
		return err
	}

	app.dbPhotos = p
	app.dbAlbums = a

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
	photos := app.dbPhotos
	albums := app.dbAlbums

	pv := make([]photoView, len(photos))
	for i, p := range photos {
		pv[i] = photoView{
			Photo:        p,
			ThumbnailURL: p.ThumbPath(),
			FullSizeURL:  p.FullPath(),
		}
	}

	av := make([]albumView, len(albums))
	for i, a := range albums {
		av[i] = albumView{
			Album:  a,
			photos: []photoView{},
		}
	}

	for i, p := range pv {
		for j, a := range av {
			if p.AlbumID == a.ID {
				av[j].photos = append(av[j].photos, pv[i])
				break
			}
		}
	}

	p := photosPage{Photos: pv, Albums: av, View: "gallery"}
	if r.URL.Query().Get("view") == "slideshow" && len(pv) > 0 {
		idx, err := strconv.Atoi(r.URL.Query().Get("index"))
		if err != nil || idx < 0 || idx >= len(pv) {
			idx = 0
		}

		p.Index = idx
		p.View = "slideshow"
		p.Current = pv[idx]
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
