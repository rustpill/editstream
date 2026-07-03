package edit

import (
	"testing"
	"time"
)

func TestParseEdit(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Edit
		wantErr bool
	}{
		{
			name: "standard edit with a byte delta",
			data: `{"type":"edit","title":"Example","user":"John","bot":false,"wiki":"enwiki","length":{"old":1000,"new":1200},"timestamp":1783114764}`,
			want: Edit{
				Wiki:      "enwiki",
				Title:     "Example",
				User:      "John",
				Type:      "edit",
				Bot:       false,
				ByteDelta: 200,
				Timestamp: time.Unix(1783114764, 0).UTC(),
			},
		},
		{
			name: "log event without length is a zero delta",
			data: `{"type":"log","title":"Special:Log","user":"Grace","bot":true,"wiki":"commonswiki","timestamp":1783114764}`,
			want: Edit{
				Wiki:      "commonswiki",
				Title:     "Special:Log",
				User:      "Grace",
				Type:      "log",
				Bot:       true,
				ByteDelta: 0,
				Timestamp: time.Unix(1783114764, 0).UTC(),
			},
		},
		{
			name:    "malformed json is an error",
			data:    `{not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEdit([]byte(tt.data))

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})

	}
}
