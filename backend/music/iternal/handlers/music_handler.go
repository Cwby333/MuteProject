package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"music/iternal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TrackInfo struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	ArtistID   string `json:"artist_id"`
	ArtistName string `json:"artist_name"`
	CoverURL   string `json:"coverUrl"`
	StreamURL  string `json:"streamUrl"`
}

func GetAllTracksHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, s3 *storage.S3Client) {
	rows, err := db.Query(r.Context(),
		`SELECT id, title, artist_id, cover_s3_key, track_s3_key FROM music`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []TrackInfo
	for rows.Next() {
		var (
			id, title, artistID string
			coverKey, trackKey  string
		)
		if err := rows.Scan(&id, &title, &artistID, &coverKey, &trackKey); err != nil {
			continue
		}

		coverURL, _ := s3.PresignGet(coverKey, 15*time.Minute)
		streamURL, _ := s3.PresignGet(trackKey, 15*time.Minute)

		list = append(list, TrackInfo{
			ID:         id,
			Title:      title,
			ArtistID:   artistID,
			ArtistName: artistID,
			CoverURL:   coverURL,
			StreamURL:  streamURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func GetUserLikedHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, s3 *storage.S3Client, userID string) {

	rows, err := db.Query(r.Context(), `
        SELECT m.id, m.title, m.artist_id, m.cover_s3_key, m.track_s3_key
        FROM music m
        JOIN liked_music l ON l.track_id = m.id
        WHERE l.user_id = $1`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []TrackInfo
	for rows.Next() {
		var id, title, artistID, coverKey, trackKey string
		rows.Scan(&id, &title, &artistID, &coverKey, &trackKey)

		coverURL, _ := s3.PresignGet(coverKey, 15*time.Minute)
		streamURL, _ := s3.PresignGet(trackKey, 15*time.Minute)

		list = append(list, TrackInfo{
			ID:         id,
			Title:      title,
			ArtistID:   artistID,
			ArtistName: artistID,
			CoverURL:   coverURL,
			StreamURL:  streamURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func UpdateTrackHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, s3Client *storage.S3Client) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Неверный URL, отсутствует track_id", http.StatusBadRequest)
		return
	}
	trackID := parts[2]

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Ошибка разбора формы: "+err.Error(), http.StatusBadRequest)
		return
	}

	var (
		setClauses []string
		args       []interface{}
		idx        = 1
	)

	if title := r.FormValue("title"); title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", idx))
		args = append(args, title)
		idx++
	}

	file, header, err := r.FormFile("cover")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Ошибка чтения cover: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err == nil {
		defer file.Close()
		buf, _ := io.ReadAll(file)
		contentType := header.Header.Get("Content-Type")
		coverKey, err := s3Client.UploadObject(buf, contentType)
		if err != nil {
			http.Error(w, "Ошибка загрузки в S3: "+err.Error(), http.StatusInternalServerError)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("cover_s3_key = $%d", idx))
		args = append(args, coverKey)
		idx++
	}

	if len(setClauses) == 0 {
		http.Error(w, "Нет полей для обновления", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf(
		"UPDATE music SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		idx,
	)
	args = append(args, trackID)

	if _, err := db.Exec(context.Background(), query, args...); err != nil {
		http.Error(w, "Ошибка обновления в базе: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func DeleteTrackHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Неверный формат URL, отсутствует track_id", http.StatusBadRequest)
		return
	}
	trackID := parts[2]
	query := `DELETE FROM music WHERE id = $1`
	_, err := db.Exec(context.Background(), query, trackID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка удаления трека из базы: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "success"}`))
}

func CreateTrackHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, s3Client *storage.S3Client) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Неверный URL, отсутствует artist_id", http.StatusBadRequest)
		return
	}
	artistID := parts[2]

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "Ошибка разбора формы: "+err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Параметр title обязателен", http.StatusBadRequest)
		return
	}

	upload := func(field, defaultCT string) (string, error) {
		file, header, err := r.FormFile(field)
		if err != nil {
			if err == http.ErrMissingFile {
				return "", fmt.Errorf("поле %q не найдено", field)
			}
			return "", err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return "", err
		}

		ct := header.Header.Get("Content-Type")
		if ct == "" {
			ct = defaultCT
		}
		return s3Client.UploadObject(data, ct)
	}

	coverKey, err := upload("cover", "image/jpeg")
	if err != nil {
		http.Error(w, "Ошибка загрузки обложки: "+err.Error(), http.StatusBadRequest)
		return
	}

	trackKey, err := upload("track", "audio/mpeg")
	if err != nil {
		http.Error(w, "Ошибка загрузки трека: "+err.Error(), http.StatusBadRequest)
		return
	}

	newID := uuid.New().String()
	createdAt := time.Now()
	query := `
        INSERT INTO music
            (id, title, artist_id, cover_s3_key, track_s3_key, created_at)
        VALUES
            ($1, $2, $3, $4, $5, $6)
    `
	if _, err := db.Exec(
		context.Background(),
		query,
		newID, title, artistID, coverKey, trackKey, createdAt,
	); err != nil {
		http.Error(w, "Ошибка записи в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success","id":"` + newID + `"}`))
}
