package server

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	Server *http.Server
}

type Config struct {
	Address         string
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func New(cfg Config, mux http.Handler) *Server {
	server := &http.Server{
		Addr:        cfg.Address,
		IdleTimeout: cfg.IdleTimeout,
		Handler:     mux,
	}
	fmt.Println(server)

	return &Server{
		Server: server,
	}
}
