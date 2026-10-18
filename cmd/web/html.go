package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"murtari.me/web/assets"
)

type htmlRenderer struct {
	templateFS      fs.FS
	sharedTemplates *template.Template
}

func assetHash(url string) (string, error) {
	const prefix = "/static/"
	if !strings.HasPrefix(url, prefix) {
		return "", fmt.Errorf("asset %q should start with %q", url, prefix)
	}

	name := strings.TrimPrefix(url, prefix)
	data, err := fs.ReadFile(assets.StaticFiles, name)
	if err != nil {
		return "", fmt.Errorf("read asset error: %w", err)
	}

	sum := sha256.Sum256(data)
	version := hex.EncodeToString(sum[:])[:8]
	return fmt.Sprintf("%s?v=%s", url, version), nil
}

func newHTMLRenderer(templateFS fs.FS, sharedTemplateFiles ...string) (*htmlRenderer, error) {
	funcs := template.FuncMap{
		"now":       time.Now,
		"assetHash": assetHash,
	}

	sharedTemplates, err := template.New("").Funcs(funcs).ParseFS(templateFS, sharedTemplateFiles...)
	if err != nil {
		return nil, err
	}

	r := &htmlRenderer{
		templateFS:      templateFS,
		sharedTemplates: sharedTemplates,
	}

	return r, nil
}

func (h *htmlRenderer) render(w http.ResponseWriter, status int, data any, templateName string, additionalTemplateFiles ...string) error {
	ts, err := h.sharedTemplates.Clone()
	if err != nil {
		return err
	}

	if len(additionalTemplateFiles) > 0 {
		ts, err = ts.ParseFS(h.templateFS, additionalTemplateFiles...)
		if err != nil {
			return err
		}
	}

	buf := new(bytes.Buffer)

	err = ts.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		return err
	}

	w.Header().Add("Vary", "HX-Request")
	w.WriteHeader(status)
	buf.WriteTo(w)

	return nil
}
