package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
)

type Request struct {
	RemoteAddr    string
	Host          string
	Method        string
	URL           string
	ContentLength int64
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := Request{
			RemoteAddr:    r.RemoteAddr,
			Host:          r.Host,
			Method:        r.Method,
			URL:           r.RequestURI,
			ContentLength: r.ContentLength,
		}

		str := fmt.Sprintf("remoteAddr: %s, host: %s, method: %s, url: %s, contentLength: %d", req.RemoteAddr, req.Host, req.Method, req.URL, req.ContentLength)

		slog.Info("logging request", slog.String("request", str))

		next.ServeHTTP(w, r)
	})
}
