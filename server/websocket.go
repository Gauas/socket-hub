package server

import (
	"context"
	"net/http"

	"github.com/Gauas/socket-hub/websocket"
)

func (s *Server) websocket(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "channel is required"})
		return
	}
	if !websocket.Valid(r) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "websocket upgrade is required"})
		return
	}

	conn, rw, ok := websocket.Upgrade(w, r)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "websocket unsupported"})
		return
	}

	socket := websocket.New(conn, s.cfg.WriteTimeout)
	client := s.hub.Subscribe(channel, socket)
	defer s.hub.Unsubscribe(client)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		defer cancel()
		_ = websocket.Read(rw.Reader)
	}()

	client.Write(ctx)
}
