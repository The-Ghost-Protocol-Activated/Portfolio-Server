package main

import (
	"log"
	"net/http"
	"os"

	"backend/db"
	"backend/handlers"

	"github.com/gorilla/mux"
)

// corsMiddleware adds the necessary headers to allow cross-origin requests.
// The allowed origin is read from the ALLOWED_ORIGIN environment variable so
// it works both locally (http://localhost:8000) and on Render (your live URL).
func corsMiddleware(next http.Handler) http.Handler {
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8000" // safe default for local dev
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Preflight request — respond immediately with 204 and no body.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database connection
	database, err := db.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize router
	r := mux.NewRouter()

	// Blog API routes — OPTIONS included on each route for preflight
	r.HandleFunc("/api/posts", handlers.GetPosts(database)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/posts/{slug}", handlers.GetPostBySlug(database)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/posts", handlers.CreatePost(database)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/posts/{id}", handlers.UpdatePost(database)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/posts/{id}", handlers.DeletePost(database)).Methods("DELETE", "OPTIONS")

	// Health check — Render can optionally ping this to confirm the service is up
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	// Wrap the entire router with CORS middleware
	handler := corsMiddleware(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
