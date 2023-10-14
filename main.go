package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Shortener is a simple URL shortener
type Shortener struct {
	mu      sync.Mutex
	counter int
	urls    map[string]string
}

// NewShortener creates a new URL shortener
func NewShortener() *Shortener {
	return &Shortener{
		urls: make(map[string]string),
	}
}

// Shorten generates a short URL for the given long URL
func (s *Shortener) Shorten(longURL string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	shortURL := base64.URLEncoding.EncodeToString([]byte(fmt.Sprint(s.counter)))

	s.urls[shortURL] = longURL
	return shortURL
}

// RedirectHandler redirects short URLs to their corresponding long URLs
func (s *Shortener) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := mux.Vars(r)["shortURL"]
	longURL, ok := s.urls[shortURL]
	if !ok {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func (s *Shortener) urlShortner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	longURL := r.FormValue("url")
	if longURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	shortURL := s.Shorten(longURL)
	fmt.Fprintf(w, "Shortened URL: http://localhost:8080/%s", shortURL)
}

func main() {
	shortener := NewShortener()

	router := mux.NewRouter()

	// Handle root endpoint
	router.HandleFunc("/{shortURL}", shortener.RedirectHandler).Methods("GET")

	// Handle /shorten endpoint
	router.HandleFunc("/shorten", shortener.urlShortner).Methods("POST")

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", router)
}
