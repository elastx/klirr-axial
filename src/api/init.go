package api

import (
	"fmt"
	"net/http"
)

func StartHTTPServer() {
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
			// Handle incoming sync requests
			w.Write([]byte("Sync complete"))
	})

	fmt.Println("HTTP server started on :8080")
	http.ListenAndServe(":8080", nil)
}
