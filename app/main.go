package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

const (
	// length of the short URL
	shortCodeLength = 6
	// use easily recognizable characters only (avoid oOl1)
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ23456789-_"
)

var (
	urlStore = make(map[string]string)
	mtx      sync.RWMutex // for locking while generating short code
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// generateShortCode generates a short unique code of length shortCodeLength
func generateShortCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// shortenHandler handles the API endpoint /shorten
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !strings.HasPrefix(req.URL, "http") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	code := generateShortCode()

	// generate a unique code
	mtx.Lock()
	for {
		if _, exists := urlStore[code]; !exists {
			urlStore[code] = req.URL
			break
		}
		code = generateShortCode()
	}
	mtx.Unlock()

	resp := ShortenResponse{
		ShortURL: "http://" + r.Host + "/" + code,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// redirectHandler handles the redirection from short URL to original URL
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	mtx.RLock()
	originalURL, ok := urlStore[code]
	mtx.RUnlock()

	if !ok {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// homeHandler serves the home page of the app
func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<body>
		<h1>URL Shortener</h1>
		<form method="POST" action="/shorten-form">
			<input type="text" name="url" placeholder="Enter a URL to shorten" required />
			<input type="submit" value="Shorten">
		</form>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// shortenFormHandler handles shorten requests through UI
func shortenFormHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	url := r.FormValue("url")
	if !strings.HasPrefix(url, "http") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode()
	mtx.Lock()
	for {
		if _, exists := urlStore[shortCode]; !exists {
			urlStore[shortCode] = url
			break
		}
		shortCode = generateShortCode()
	}
	mtx.Unlock()

	shortURL := "http://" + r.Host + "/" + shortCode

	responseHTML := `
	<!DOCTYPE html>
	<html>
	<body>
		<p>Short URL: <a href="` + shortURL + `">` + shortURL + `</a></p>
		<a href="/">Home</a>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(responseHTML))
}

// getPort
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

// authMiddleware handles Basic Authentication
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := os.Getenv("AUTH_USER")
		password := os.Getenv("AUTH_PASS")

		u, p, ok := r.BasicAuth()
		if !ok || u != username || p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// as of go 1.20, no need to call Seed()
	// rand.Seed(time.Now().UnixNano())
	r := mux.NewRouter()

	// Public routes (UI and redirection)
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/{code}", redirectHandler).Methods("GET")

	// Protected routes
	restricted := r.NewRoute().Subrouter()
	restricted.Use(authMiddleware)
	restricted.HandleFunc("/shorten", shortenHandler).Methods("POST")
	restricted.HandleFunc("/shorten-form", shortenFormHandler).Methods("POST")

	addr := getPort()
	log.Printf("URL Shortener service running on %s\n", addr)
	// Exit logging errors, if any
	log.Fatal(http.ListenAndServe(addr, r))
}
