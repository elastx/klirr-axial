package api

import (
	"fmt"
	"net/http"
)

func StartHTTPServer() {
	http.HandleFunc("/v1/sync", handleSync)

	fmt.Println("HTTP server started on :8080")
	http.ListenAndServe(":8080", nil)
}

