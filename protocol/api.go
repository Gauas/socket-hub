package protocol

import (
	"encoding/json"
	"fmt"
)

const Publish = "publish"

type Command struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type PublishParams struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type Reply struct {
	Error  *Error `json:"error"`
	Result any    `json:"result,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %d", e.Message, e.Code)
}

func Fail(code int, message string) Reply {
	return Reply{Error: &Error{Code: code, Message: message}}
}

func Decode(raw []byte) (Command, error) {
	var command Command
	err := json.Unmarshal(raw, &command)
	return command, err
}

func (c Command) Publish() (PublishParams, error) {
	var params PublishParams
	err := json.Unmarshal(c.Params, &params)
	if len(params.Data) == 0 {
		params.Data = json.RawMessage("{}")
	}
	return params, err
}
