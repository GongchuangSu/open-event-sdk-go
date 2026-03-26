package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEventFilter(t *testing.T) {
	tests := []struct {
		name       string
		eventsFlag string
		wantAll    bool
		wantCodes  []string
	}{
		{
			name:       "empty flag matches all",
			eventsFlag: "",
			wantAll:    true,
		},
		{
			name:       "single event code",
			eventsFlag: "kso.app_chat.message.create",
			wantAll:    false,
			wantCodes:  []string{"kso.app_chat.message.create"},
		},
		{
			name:       "multiple event codes",
			eventsFlag: "kso.app_chat.message.create,kso.app_chat.create",
			wantAll:    false,
			wantCodes:  []string{"kso.app_chat.message.create", "kso.app_chat.create"},
		},
		{
			name:       "whitespace trimmed",
			eventsFlag: " kso.a , kso.b , kso.c ",
			wantAll:    false,
			wantCodes:  []string{"kso.a", "kso.b", "kso.c"},
		},
		{
			name:       "only commas treated as empty",
			eventsFlag: ",,",
			wantAll:    true,
		},
		{
			name:       "trailing comma ignored",
			eventsFlag: "kso.a,",
			wantAll:    false,
			wantCodes:  []string{"kso.a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewEventFilter(tt.eventsFlag)
			assert.Equal(t, tt.wantAll, f.all)
			if !tt.wantAll {
				for _, code := range tt.wantCodes {
					_, exists := f.codes[code]
					assert.True(t, exists, "expected code %q in filter", code)
				}
			}
		})
	}
}

func TestEventFilter_Match(t *testing.T) {
	tests := []struct {
		name       string
		eventsFlag string
		eventCode  string
		want       bool
	}{
		{
			name:       "match all when no filter",
			eventsFlag: "",
			eventCode:  "kso.any.event",
			want:       true,
		},
		{
			name:       "match exact code",
			eventsFlag: "kso.app_chat.message.create",
			eventCode:  "kso.app_chat.message.create",
			want:       true,
		},
		{
			name:       "no match for different code",
			eventsFlag: "kso.app_chat.message.create",
			eventCode:  "kso.app_chat.create",
			want:       false,
		},
		{
			name:       "match one of multiple codes",
			eventsFlag: "kso.a,kso.b,kso.c",
			eventCode:  "kso.b",
			want:       true,
		},
		{
			name:       "no partial match",
			eventsFlag: "kso.app_chat",
			eventCode:  "kso.app_chat.message.create",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewEventFilter(tt.eventsFlag)
			assert.Equal(t, tt.want, f.Match(tt.eventCode))
		})
	}
}
