package http

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Gauas/socket-hub/protocol"
	"github.com/Gauas/socket-hub/realtime"
)

func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if !s.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	scanner := bufio.NewScanner(r.Body)
	encoder := json.NewEncoder(w)

	for scanner.Scan() {
		cmd, err := protocol.Decode(scanner.Bytes())
		if err != nil {
			_ = encoder.Encode(protocol.Fail(400, "invalid command"))
			continue
		}

		_ = encoder.Encode(s.handle(r, cmd))
	}
}

func (s *Server) handle(r *http.Request, cmd protocol.Command) protocol.Reply {
	if cmd.Method != protocol.Publish {
		return protocol.Fail(400, "unsupported method")
	}

	params, err := cmd.Publish()
	if err != nil {
		return protocol.Fail(400, "invalid publish params")
	}
	if params.Channel == "" {
		return protocol.Fail(400, "channel is required")
	}

	result, err := s.hub.Publish(r.Context(), realtime.Publication{
		Channel: params.Channel,
		Data:    params.Data,
	})
	if err != nil {
		return protocol.Fail(500, err.Error())
	}

	return protocol.Reply{Result: result}
}

func (s *Server) authorized(r *http.Request) bool {
	if s.cfg.APIKey == "" {
		return true
	}

	header := r.Header.Get("Authorization")
	return strings.TrimSpace(header) == "apikey "+s.cfg.APIKey
}
