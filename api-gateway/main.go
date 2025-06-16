package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var (
	usersServiceURL = getEnv("USERS_SERVICE_URL", "http://users-service:8888")
	musicServiceURL = getEnv("MUSIC_SERVICE_URL", "http://music-service:8080")
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// Append original query parameters
	if r.URL.RawQuery != "" {
		targetURL = targetURL + "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		if strings.HasPrefix(strings.ToLower(name), "access-control-") {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	fmt.Printf("usersServiceURL: %v\n", usersServiceURL)
	fmt.Printf("musicServiceURL: %v\n", musicServiceURL)

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "API Gateway running")
	}).Methods("GET")

	// Users Service
	router.HandleFunc("/user/register", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/register", usersServiceURL)
		fmt.Println(targetURL)
		proxyRequest(w, r, targetURL)
	}).Methods("POST", "OPTIONS")

	router.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/login", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("POST")

	router.HandleFunc("/user/logout", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/logout", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("POST")

	router.HandleFunc("/user/refresh", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/refresh", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("POST")

	router.HandleFunc("/user/get", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/get", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("GET")

	router.HandleFunc("/user/all", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/all", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("GET")

	router.HandleFunc("/user/delete", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/delete", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("DELETE")

	router.HandleFunc("/user/update", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/update", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("PUT")

	router.HandleFunc("/user/track/favorite", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/track/favorite", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("POST")

	router.HandleFunc("/user/track/favorite", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/user/track/favorite", usersServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("DELETE")

	// Music Service
	router.HandleFunc("/tracks", func(w http.ResponseWriter, r *http.Request) {
		targetURL := fmt.Sprintf("%s/tracks", musicServiceURL)
		proxyRequest(w, r, targetURL)
	}).Methods("GET")

	router.HandleFunc("/tracks/{userId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["userId"]
		targetURL := fmt.Sprintf("%s/tracks/%s", musicServiceURL, userId)
		proxyRequest(w, r, targetURL)
	}).Methods("GET")

	router.HandleFunc("/track/{trackId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		trackId := vars["trackId"]
		targetURL := fmt.Sprintf("%s/track/%s", musicServiceURL, trackId)
		proxyRequest(w, r, targetURL)
	}).Methods("POST")

	router.HandleFunc("/track/{trackId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		trackId := vars["trackId"]
		targetURL := fmt.Sprintf("%s/track/%s", musicServiceURL, trackId)
		proxyRequest(w, r, targetURL)
	}).Methods("DELETE")

	router.HandleFunc("/track/{userId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["userId"]
		targetURL := fmt.Sprintf("%s/track/%s", musicServiceURL, userId)
		proxyRequest(w, r, targetURL)
	}).Methods("POST")

	// Add simple stream endpoint for player functionality
	router.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Stream not available")) // Заглушка
	}).Methods("GET")

	router.HandleFunc("/play/{trackName}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "playing", "message": "Player functionality not implemented yet"}`)) // Заглушка
	}).Methods("GET")

	router.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "paused", "message": "Player functionality not implemented yet"}`)) // Заглушка
	}).Methods("GET")

	router.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "resumed", "message": "Player functionality not implemented yet"}`)) // Заглушка
	}).Methods("GET")

	router.HandleFunc("/user/{userId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["userId"]
		targetURL := fmt.Sprintf("%s/user/%s", usersServiceURL, userId)
		proxyRequest(w, r, targetURL)
	}).Methods("GET")

	// Configure CORS with more permissive settings for development
	corsOptions := cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
	corsHandler := cors.New(corsOptions).Handler(router)

	log.Printf("Gateway running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", corsHandler))
}
