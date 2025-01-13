package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DataBlock struct {
	ID      string `json:"id"`
	User    string `json:"user"`
	Content string `json:"content"`
}

var localData []DataBlock

func StartHTTPServer() {
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var requestData []DataBlock
			if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			// Process received data and update local state
			for _, block := range requestData {
				localData = append(localData, block)
			}
			w.Write([]byte("Sync complete"))
		} else if r.Method == http.MethodGet {
			// Serve local data
			json.NewEncoder(w).Encode(localData)
		}
	})

	fmt.Println("HTTP server started on :8080")
	http.ListenAndServe(":8080", nil)
}

func SetData(data []DataBlock) {
	localData = data
}
