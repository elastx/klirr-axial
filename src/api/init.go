package api

import (
	"fmt"
	"net/http"
)

func StartHTTPServer() {
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		// if r.Method == http.MethodPost {
		// 	var requestData []data.DataBlock
		// 	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		// 		http.Error(w, "Invalid request", http.StatusBadRequest)
		// 		return
		// 	}

		// 	// Process received data and update local state
		// 	localData = append(localData, requestData...)
		// 	w.Write([]byte("Sync complete"))
		// } else if r.Method == http.MethodGet {
		// 	// Serve local data
		// 	json.NewEncoder(w).Encode(localData)
		// }
	})

	fmt.Println("HTTP server started on :8080")
	http.ListenAndServe(":8080", nil)
}

