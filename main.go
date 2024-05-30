package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Shortener is a simple URL shortener
type Shortener struct {
	mu         sync.Mutex
	counter    int
	urls       map[string]string
	collection *mongo.Collection
}

// NewShortener creates a new URL shortener and initializes MongoDB connection
func NewShortener() *Shortener {
	// clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	clientOptions := options.Client().ApplyURI("mongodb+srv://saipavank1107:8LOt0HsRwiYwIfQJ@cluster0.n7pipl5.mongodb.net/?retryWrites=true&w=majority")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	database := client.Database("urlshortener")
	collection := database.Collection("urls")

	return &Shortener{
		urls:       make(map[string]string),
		collection: collection,
	}
}

// Shorten generates a short URL for the given long URL and stores it in MongoDB
func (s *Shortener) Shorten(longURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	shortURL := base64.URLEncoding.EncodeToString([]byte(fmt.Sprint(s.counter)))

	_, err := s.collection.InsertOne(context.Background(), map[string]interface{}{
		"shortURL": shortURL,
		"longURL":  longURL,
	})
	if err != nil {
		return "", err
	}

	s.urls[shortURL] = longURL
	return shortURL, nil
}

// RedirectHandler redirects short URLs to their corresponding long URLs
func (s *Shortener) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := mux.Vars(r)["shortURL"]
	longURL, ok := s.urls[shortURL]
	if !ok {
		// Check MongoDB for the short URL
		var result struct {
			LongURL string `bson:"longURL"`
		}
		filter := map[string]interface{}{"shortURL": shortURL}
		err := s.collection.FindOne(context.Background(), filter).Decode(&result)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		longURL = result.LongURL
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

	shortURL, err := s.Shorten(longURL)
	if err != nil {
		http.Error(w, "Error creating short URL", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Shortened URL: http://localhost:8080/%s", shortURL)
}

func (s *Shortener) home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var a string
	a = "HELLO WORLD..."

	fmt.Fprintf(w, "/%s", a)
}

func main() {
	shortener := NewShortener()

	router := mux.NewRouter()

	// Handle root endpoint
	router.HandleFunc("/{shortURL}", shortener.RedirectHandler).Methods("GET")

	router.HandleFunc("/", shortener.home).Methods("GET")

	// Handle /shorten endpoint
	router.HandleFunc("/shorten", shortener.urlShortner).Methods("POST")

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", router)
}
