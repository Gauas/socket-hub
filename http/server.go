package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Gauas/socket-hub/config"
	"github.com/Gauas/socket-hub/realtime"
)

type Server struct {
	cfg *config.Config
	hub *realtime.Hub
}

func New(cfg *config.Config, hub *realtime.Hub) *Server {
	return &Server{cfg: cfg, hub: hub}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/api", s.api)
	mux.HandleFunc("/connection/websocket", s.websocket)
	mux.HandleFunc("/ws", s.websocket)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", s.cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("socket-hub shutdown error: %v", err)
		}
	}()

	log.Printf("socket-hub listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
