package frontend

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed dist
var distFS embed.FS

// Handler returns an http.Handler that serves the frontend static files
func Handler() http.Handler {
	// Check if we're in development mode (dist directory exists)
	if _, err := os.Stat("src/frontend/dist"); err == nil {
		fmt.Println("Development mode: serving frontend from disk")
		return http.FileServer(http.Dir("src/frontend/dist"))
	}

	fmt.Println("Production mode: serving embedded frontend files")
	fsys := fs.FS(distFS)
	contentFS, err := fs.Sub(fsys, "dist")
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request URL:", r.URL.Path)	
		// Clean the path by removing leading slash
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// First try to serve the file directly
		if file, err := contentFS.Open(path); err == nil {
			defer file.Close()
			http.FileServer(http.FS(contentFS)).ServeHTTP(w, r)
			return
		}

		// If file not found, serve index.html for client-side routing
		indexBytes, err := fs.ReadFile(contentFS, "index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		http.ServeContent(w, r, "index.html", time.Now(), newInMemoryFile(indexBytes))
	})
}

// inMemoryFile implements io.ReadSeeker for serving files from memory
type inMemoryFile struct {
	*bytes.Reader
}

func newInMemoryFile(data []byte) *inMemoryFile {
	return &inMemoryFile{bytes.NewReader(data)}
} 