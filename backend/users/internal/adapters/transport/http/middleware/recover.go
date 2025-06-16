package middleware

import (
	gojson "github.com/goccy/go-json"
	"log/slog"
	"net/http"

	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/lib"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				slog.Info("recover middleware", slog.Any("recover", r))

				resp := lib.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "server error",
				}
				data, err := gojson.Marshal(resp)
				if err != nil {
					slog.Info("recover middle json marshal", slog.String("error", err.Error()))

					http.Error(w, "server error", http.StatusInternalServerError)
					return
				}

				http.Error(w, string(data), http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}
