package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	http.HandleFunc("/shorten", shortenHandler(db))
	http.HandleFunc("/", getShortenHandler(db))

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
