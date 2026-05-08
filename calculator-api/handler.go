package main

import (
	"encoding/json"
	"net/http"
)

type Request struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type Response struct {
	Result float64 `json:"result"`
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	result := req.A + req.B

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: result})
}

func subtractHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	result := req.A - req.B

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: result})
}

func multiplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	result := req.A * req.B

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: result})
}

func divideHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.B == 0 {
		http.Error(w, "division by zero", http.StatusBadRequest)
	}

	result := req.A * req.B

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: result})
}
