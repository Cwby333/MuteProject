package musicserver

import (
	"net"
	"net/http"
	"strings"

	"music/iternal/handlers"
	"music/iternal/storage"

	"github.com/jackc/pgx/v5"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT, POST, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setupRoutes(db *pgx.Conn, s3Client *storage.S3Client) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/track/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateTrackHandler(w, r, db, s3Client)
		case http.MethodPatch:
			handlers.UpdateTrackHandler(w, r, db, s3Client)
		case http.MethodDelete:
			handlers.DeleteTrackHandler(w, r, db)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/tracks/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		userID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/tracks/"), "/")

		if userID == "" {
			handlers.GetAllTracksHandler(w, r, db, s3Client)
		} else {
			handlers.GetUserLikedHandler(w, r, db, s3Client, userID)
		}
	})

	return corsMiddleware(mux)
}

func StartServer(address string, db *pgx.Conn, s3Client *storage.S3Client) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := &http.Server{
		Handler: setupRoutes(db, s3Client),
	}
	return server.Serve(lis)
}
