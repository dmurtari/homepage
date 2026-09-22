package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"time"
)

type htmlRenderer struct {
	templateFS      fs.FS
	sharedTemplates *template.Template
}

func hashPath(path string) string {
	data, _ := os.ReadFile(path)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:8]
}

func assetHash(path string) string {
	return fmt.Sprintf("%s?v=%s", path, hashPath(path))
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
