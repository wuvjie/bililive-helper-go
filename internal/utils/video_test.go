package utils

import (
	"testing"
	"time"
)

func TestParseFilename_Normal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantOk    bool
		wantYear  int
		wantMonth time.Month
		wantDay   int
		wantHour  int
		wantMin   int
		wantSec   int
	}{
		{
			name:      "standard flv file",
			input:     "[2026-06-04 00-19-00][怡宝][怡宝]001.flv",
			wantKey:   "[怡宝][怡宝]",
			wantOk:    true,
			wantYear:  2026,
			wantMonth: time.June,
			wantDay:   4,
			wantHour:  0,
			wantMin:   19,
			wantSec:   0,
		},
		{
			name:      "standard mp4 file",
			input:     "[2025-12-31 23-59-59][streamer][title]002.mp4",
			wantKey:   "[streamer][title]",
			wantOk:    true,
			wantYear:  2025,
			wantMonth: time.December,
			wantDay:   31,
			wantHour:  23,
			wantMin:   59,
			wantSec:   59,
		},
		{
			name:      "ts file",
			input:     "[2026-01-15 10-30-00][主播名][房间标题]003.ts",
			wantKey:   "[主播名][房间标题]",
			wantOk:    true,
			wantYear:  2026,
			wantMonth: time.January,
			wantDay:   15,
			wantHour:  10,
			wantMin:   30,
			wantSec:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, dt, ok := ParseFilename(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("ParseFilename(%q): ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if key != tt.wantKey {
				t.Errorf("ParseFilename(%q): key = %q, want %q", tt.input, key, tt.wantKey)
			}
			if dt.Year() != tt.wantYear || dt.Month() != tt.wantMonth || dt.Day() != tt.wantDay {
				t.Errorf("ParseFilename(%q): date = %v, want %d-%02d-%02d", tt.input, dt, tt.wantYear, tt.wantMonth, tt.wantDay)
			}
			if dt.Hour() != tt.wantHour || dt.Minute() != tt.wantMin || dt.Second() != tt.wantSec {
				t.Errorf("ParseFilename(%q): time = %02d:%02d:%02d, want %02d:%02d:%02d", tt.input,
					dt.Hour(), dt.Minute(), dt.Second(), tt.wantHour, tt.wantMin, tt.wantSec)
			}
		})
	}
}

func TestParseFilename_MergedFile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantOk  bool
	}{
		{
			name:    "merged mp4 file",
			input:   "[2026-06-03 21-06-44][怡宝][怡宝]-合并版.mp4",
			wantKey: "[怡宝][怡宝]",
			wantOk:  true,
		},
		{
			name:    "merged flv file",
			input:   "[2026-06-03 21-06-44][怡宝][怡宝]-合并版.flv",
			wantKey: "[怡宝][怡宝]",
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _, ok := ParseFilename(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("ParseFilename(%q): ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if key != tt.wantKey {
				t.Errorf("ParseFilename(%q): key = %q, want %q", tt.input, key, tt.wantKey)
			}
		})
	}
}

func TestParseFilename_NoMatch(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "plain text no brackets",
			input: "randomvideo.mp4",
		},
		{
			name:  "missing time part",
			input: "[streamer][title]001.mp4",
		},
		{
			name:  "wrong date format",
			input: "[2026/06/04 00-19-00][怡宝][怡宝]001.mp4",
		},
		{
			name:  "unsupported extension",
			input: "[2026-06-04 00-19-00][怡宝][怡宝]001.mkv",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "single bracket segment",
			input: "[2026-06-04 00-19-00][怡宝]001.mp4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, ok := ParseFilename(tt.input)
			if ok {
				t.Errorf("ParseFilename(%q): expected no match, but got ok=true", tt.input)
			}
		})
	}
}

func TestParseFilename_DifferentExtensions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantKey string
	}{
		{
			name:    "flv extension",
			input:   "[2026-06-04 00-19-00][怡宝][怡宝]001.flv",
			wantOk:  true,
			wantKey: "[怡宝][怡宝]",
		},
		{
			name:    "mp4 extension",
			input:   "[2026-06-04 00-19-00][怡宝][怡宝]001.mp4",
			wantOk:  true,
			wantKey: "[怡宝][怡宝]",
		},
		{
			name:    "ts extension",
			input:   "[2026-06-04 00-19-00][怡宝][怡宝]001.ts",
			wantOk:  true,
			wantKey: "[怡宝][怡宝]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _, ok := ParseFilename(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("ParseFilename(%q): ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && key != tt.wantKey {
				t.Errorf("ParseFilename(%q): key = %q, want %q", tt.input, key, tt.wantKey)
			}
		})
	}
}
