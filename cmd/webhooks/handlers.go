package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
)

const maxBodyBytes = 1 << 20 // 1 MB

func (app *application) activities(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.logger.Error(err.Error())
		var tooBig *http.MaxBytesError

		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	apiKey := r.Header.Get("X-Api-Key")
	if subtle.ConstantTimeCompare([]byte(apiKey), []byte(app.apiKey)) != 1 {
		app.logger.Error("invalid API key")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	signature, err := hex.DecodeString(r.Header.Get("X-Signature"))
	if err != nil {
		app.logger.Error("missing signature header")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !app.verifySignature([]byte(signature), body) {
		app.logger.Error("invalid signature")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (app *application) verifySignature(signature, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(app.secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(signature, expectedMAC)
}
