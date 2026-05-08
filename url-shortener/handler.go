package main

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"net/url"

	"gorm.io/gorm" // ← was bolt
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLen = 6

func generateCode() (string, error) {
	code := make([]byte, codeLen)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

func shortenHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if _, err := url.ParseRequestURI(req.URL); err != nil {
			http.Error(w, "invalid url", http.StatusUnprocessableEntity)
			return
		}

		code, err := generateCode()
		if err != nil {
			http.Error(w, "could not generate code", http.StatusInternalServerError)
			return
		}

		if err := saveLink(db, code, req.URL); err != nil {
			http.Error(w, "could not save link", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shortenResponse{
			Code:     code,
			ShortURL: "http://localhost:8080/" + code,
		})
	}
}
