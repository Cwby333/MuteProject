package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/lib"
	gojson "github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
)

func AccessJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		jwtCookie, err := r.Cookie("jwt-access")
		if err == nil && token == "" {
			token = jwtCookie.Value
		}

		if token == "" {
			slog.Info("missing auth token")

			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "missing auth token",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "missing auth token", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		secretKey := os.Getenv("JWT_SECRET_KEY")

		t, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		}, jwt.WithIssuer(os.Getenv("JWT_ISSUER")),
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
			jwt.WithExpirationRequired(),
		)

		if err != nil {
			slog.Info("jwt parse", slog.String("error", err.Error()))

			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		if !t.Valid {
			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			resp := lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "serve error", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		typeToken, ok := claims["type"].(string)
		if !ok {
			resp := lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "serve error", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		if typeToken != "access" {
			resp := lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "wrong token type",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong token type", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "claims", claims)
		r = r.WithContext(ctx)

		slog.Info("success jwt-access middleware")

		next.ServeHTTP(w, r)
	})
}

func RefreshJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//	TODO: Maybe change
		//	token := r.Header.Get("Authorization")
		//	token = strings.TrimPrefix(token, "Bearer ")
		//	TO Cookie["jwt-refresh"]

		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		jwtCookie, err := r.Cookie("jwt-refresh")
		if err == nil && token == "" {
			token = jwtCookie.Value
		}

		if token == "" {
			slog.Info("missing refresh token")

			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "missing auth token",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "missing auth token", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		secretKey := os.Getenv("JWT_SECRET_KEY")

		t, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		}, jwt.WithIssuer(os.Getenv("JWT_ISSUER")),
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
			jwt.WithExpirationRequired(),
		)

		if err != nil {
			slog.Info("jwt parse", slog.String("error", err.Error()))

			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		if !t.Valid {
			slog.Info("jwt invalid")

			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			resp := lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "serve error", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		typeToken, ok := claims["type"].(string)
		if !ok {
			resp := lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "serve error", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		if typeToken != "refresh" {
			resp := lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "wrong token type",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong token type", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(data), http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "claims", claims)
		r = r.WithContext(ctx)

		slog.Info("success jwt-refresh middleware")

		next.ServeHTTP(w, r)
	})
}
