package api

import (
	"fmt"
	"net/http"

	"axial/frontend"
)

func StartHTTPServer() {
	// API endpoints
	http.HandleFunc("/v1/sync", handleSync)

	// Frontend serving
	http.Handle("/", frontend.Handler())

	fmt.Println("HTTP server started on :8080")
	http.ListenAndServe(":8080", nil)
}

