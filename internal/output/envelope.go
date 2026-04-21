// Package output formats lnav-cli stdout data and stderr errors.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Meta is attached to every JSON response as "_meta".
type Meta struct {
	Source string `json:"source,omitempty"`
	Since  string `json:"since,omitempty"`
	Until  string `json:"until,omitempty"`
	Count  int    `json:"count"`
}

type envelope struct {
	Data any  `json:"data"`
	Meta Meta `json:"_meta"`
}

// WriteJSON writes a single JSON envelope containing all rows and metadata.
func WriteJSON(w io.Writer, rows []map[string]any, meta Meta) error {
	meta.Count = len(rows)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(envelope{Data: rows, Meta: meta})
}

// WriteNDJSON writes one JSON object per line (no envelope) for streaming consumers.
func WriteNDJSON(w io.Writer, rows []map[string]any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

// Err is the structured error envelope written to stderr.
type Err struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Hint       string `json:"hint,omitempty"`
	ConsoleURL string `json:"console_url,omitempty"`
}

func (e *Err) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// Errorf builds a new structured error.
func Errorf(code, format string, a ...any) *Err {
	return &Err{Code: code, Message: fmt.Sprintf(format, a...)}
}

// WithHint attaches a human-actionable hint to the error.
func (e *Err) WithHint(hint string) *Err { e.Hint = hint; return e }

// WriteErr emits the error as JSON to stderr.
func WriteErr(w io.Writer, e *Err) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}
