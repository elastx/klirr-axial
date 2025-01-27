package api

import (
	"log"
	"net/http"
	"os"
	"strings"
)

type spaFileSystem struct {
	root    http.FileSystem
	indexes bool
}

func (fs *spaFileSystem) Open(name string) (http.File, error) {
	log.Printf("Attempting to serve: %s", name)
	
	// Don't interfere with API routes
	if strings.HasPrefix(name, "/v1/") {
		return nil, os.ErrNotExist
	}

	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		// Serve index.html for any path that doesn't exist
		return fs.root.Open("index.html")
	}
	return f, err
}

func RegisterRoutes() {
	// Log current working directory
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s", cwd)

	// Serve frontend files with SPA support
	fs := &spaFileSystem{root: http.Dir("src/frontend/dist"), indexes: true}
	http.Handle("/", http.FileServer(fs))

	http.HandleFunc("/v1/ping", handlePing)
	http.HandleFunc("/v1/sync", handleSync)
	http.HandleFunc("/v1/sync/messages", handleSyncMessages)
	http.HandleFunc("/v1/sync/users", handleSyncUsers)

	// User routes
	http.HandleFunc("/v1/users", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Users endpoint: %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			handleGetUsers(w, r)
		case http.MethodPost:
			handleRegisterUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Message routes
	http.HandleFunc("/v1/messages", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Messages endpoint: %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			handleGetMessages(w, r)
		case http.MethodPost:
			handleCreateMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Topics route
	http.HandleFunc("/v1/topics", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Topics endpoint: %s %s", r.Method, r.URL.Path)
		if r.Method == http.MethodGet {
			handleGetTopics(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	
	log.Printf("Routes registered: / (SPA + static files), /v1/users, /v1/messages, /v1/topics")
} 