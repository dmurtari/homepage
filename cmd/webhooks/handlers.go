package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"io"
	"net/http"
)

func (app *application) workout(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	signature := r.Header.Get("X-Signature")
	if signature == "" {
		app.logger.Error("missing signature header")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !app.verifySignature([]byte(signature), body) {
		app.logger.Error("invalid signature")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *application) verifySignature(signature, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(app.secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(signature, expectedMAC)
}
