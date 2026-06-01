package api

import (
	"log"
	"net/http"
	"os"
	"strings"

	"gorm.io/gorm"
)

// Server holds the per-process state the HTTP handlers need — currently
// just the *gorm.DB. Creating one binds the routes to a specific DB, so
// tests can spin up an in-memory SQLite and exercise handlers directly.
type Server struct {
	DB *gorm.DB
}

func NewServer(db *gorm.DB) *Server {
	return &Server{DB: db}
}

type spaFileSystem struct {
	root    http.FileSystem
	indexes bool
}

func (fs *spaFileSystem) Open(name string) (http.File, error) {
	log.Printf("Attempting to serve: %s", name)

	if strings.HasPrefix(name, "/v1/") {
		return nil, os.ErrNotExist
	}

	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("index.html")
	}
	return f, err
}

func (s *Server) RegisterRoutes() {
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s", cwd)

	fs := &spaFileSystem{root: http.Dir("frontend/dist"), indexes: true}
	http.Handle("/", http.FileServer(fs))

	http.HandleFunc("/v1/ping", s.handlePing)
	http.HandleFunc("/v1/sync", s.handleSync)
	http.HandleFunc("/v1/sync/messages", s.handleSyncMessages)
	http.HandleFunc("/v1/sync/bulletins", s.handleSyncBulletins)
	http.HandleFunc("/v1/sync/users", s.handleSyncUsers)

	// User routes
	http.HandleFunc("/v1/users/search", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Users search endpoint: %s %s", r.Method, r.URL.Path)
		if r.Method == http.MethodGet {
			s.handleSearchUsers(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/v1/users/recent", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Users recent endpoint: %s %s", r.Method, r.URL.Path)
		if r.Method == http.MethodGet {
			s.handleRecentUsers(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/v1/users/{fingerprint}", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("User endpoint: %s %s", r.Method, r.URL.Path)
		if r.Method == http.MethodGet {
			s.handleGetUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/v1/users", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Users endpoint: %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			s.handleGetUsers(w, r)
		case http.MethodPost:
			s.handleRegisterUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Message routes
	http.HandleFunc("/v1/messages", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Messages endpoint: %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			s.handleGetMessages(w, r)
		case http.MethodPost:
			s.handleCreateMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Bulletin routes
	http.HandleFunc("/v1/bulletin", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Bulletin endpoint: %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			s.handleGetBulletin(w, r)
		case http.MethodPost:
			s.handleCreateBulletin(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}
