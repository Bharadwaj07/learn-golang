package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/subtract", subtractHandler)
	http.HandleFunc("/multiply", multiplyHandler)
	http.HandleFunc("/divide", multiplyHandler)
	fmt.Println("Server running on http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
