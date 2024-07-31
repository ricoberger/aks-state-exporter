package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Address string `json:"address" env:"ADDRESS" default:":8080" help:"The address where the server should listen on."`
}

// Server is the interface of a server, which provides the options to start and
// stop the underlying http server.
type Server interface {
	Start()
	Stop()
}

// server implements the Server interface.
type server struct {
	server *http.Server
}

// Start starts serving the server.
func (s *server) Start() {
	slog.Info("Server started", slog.String("address", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			slog.Error("Server died unexpected", slog.String("error", err.Error()))
		}
	}
}

// Stop terminates the server gracefully.
func (s *server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.DebugContext(ctx, "Start shutdown of the server")

	err := s.server.Shutdown(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Graceful shutdown of the server failed", slog.String("error", err.Error()))
	}
}

// New return a new server. It creates the underlying http server, with the
// given address and all the required routes mounted.
func New(config Config) (Server, error) {
	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	router.Handle("/metrics", promhttp.Handler())

	return &server{
		server: &http.Server{
			Addr:              config.Address,
			Handler:           router,
			ReadHeaderTimeout: 3 * time.Second,
		},
	}, nil
}
