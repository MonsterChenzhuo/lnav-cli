package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteJSON_Envelope(t *testing.T) {
	buf := &bytes.Buffer{}
	err := WriteJSON(buf, []map[string]any{{"ts": "2026-04-21T00:00:00Z", "body": "hello"}}, Meta{Source: "nginx-prod"})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	var got struct {
		Data []map[string]any `json:"data"`
		Meta map[string]any   `json:"_meta"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v\nraw=%s", err, buf.String())
	}
	if len(got.Data) != 1 || got.Data[0]["body"] != "hello" {
		t.Fatalf("data wrong: %+v", got.Data)
	}
	if got.Meta["source"] != "nginx-prod" {
		t.Fatalf("meta wrong: %+v", got.Meta)
	}
}

func TestWriteNDJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	err := WriteNDJSON(buf, []map[string]any{{"a": 1}, {"a": 2}})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %q", len(lines), buf.String())
	}
}

func TestErr_JSONShape(t *testing.T) {
	e := Errorf("lnav_not_found", "lnav executable not found on PATH").WithHint("run: lnav-cli setup")
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), "\"code\":\"lnav_not_found\"") ||
		!strings.Contains(string(raw), "\"hint\":\"run: lnav-cli setup\"") {
		t.Fatalf("shape wrong: %s", raw)
	}
	if e.Error() == "" {
		t.Fatal("Error() empty")
	}
}
