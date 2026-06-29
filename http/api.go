package http

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Gauas/socket-hub/realtime"
)

type command struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type publishParams struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type reply struct {
	Error  *apiError `json:"error"`
	Result any       `json:"result,omitempty"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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
		var cmd command
		if err := json.Unmarshal(scanner.Bytes(), &cmd); err != nil {
			_ = encoder.Encode(reply{Error: &apiError{Code: 400, Message: "invalid command"}})
			continue
		}

		_ = encoder.Encode(s.handle(r, cmd))
	}
}

func (s *Server) handle(r *http.Request, cmd command) reply {
	if cmd.Method != "publish" {
		return reply{Error: &apiError{Code: 400, Message: "unsupported method"}}
	}

	var params publishParams
	if err := json.Unmarshal(cmd.Params, &params); err != nil {
		return reply{Error: &apiError{Code: 400, Message: "invalid publish params"}}
	}
	if params.Channel == "" {
		return reply{Error: &apiError{Code: 400, Message: "channel is required"}}
	}
	if len(params.Data) == 0 {
		params.Data = json.RawMessage("{}")
	}

	result, err := s.hub.Publish(r.Context(), realtime.Publication{
		Channel: params.Channel,
		Data:    params.Data,
	})
	if err != nil {
		return reply{Error: &apiError{Code: 500, Message: err.Error()}}
	}

	return reply{Result: result}
}

func (s *Server) authorized(r *http.Request) bool {
	if s.cfg.APIKey == "" {
		return true
	}

	header := r.Header.Get("Authorization")
	return strings.TrimSpace(header) == "apikey "+s.cfg.APIKey
}
