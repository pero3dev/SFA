package handlers

import (
	"encoding/json"
	"io"
)

func jsonEncoder(w io.Writer, payload any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	return encoder.Encode(payload)
}
