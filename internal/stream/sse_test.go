package stream

import (
	"bytes"
	"testing"
)

func TestDataPayload(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
		ok   bool
	}{
		{"data line with space", `data: {"type":"edit"}`, `{"type":"edit"}`, true},
		{"data line without space", `data:{"type":"edit"}`, `{"type":"edit"}`, true},
		{"event line is skipped", `event: message`, "", false},
		{"id line is skipped", `id: [{"topic":"x"}]`, "", false},
		{"comment line is skipped", `: heartbeat`, "", false},
		{"empty line is skipped", ``, "", false},
		{"empty data line is skipped", `data: `, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := dataPayload([]byte(tt.line))
			if ok != tt.ok {
				t.Fatalf("ok %v, want %v", ok, tt.ok)
			}
			if ok && !bytes.Equal(got, []byte(tt.want)) {
				t.Errorf("payload = %q, want %q", got, tt.want)
			}
		})
	}
}
