package utils

import "testing"

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{name: "zero bytes", bytes: 0, expected: "0.00 KB"},
		{name: "hundreds of bytes", bytes: 512, expected: "0.50 KB"},
		{name: "exactly 1 KB", bytes: 1024, expected: "1.00 KB"},
		{name: "1.5 KB", bytes: 1536, expected: "1.50 KB"},
		{name: "exactly 1 MB", bytes: 1024 * 1024, expected: "1.00 MB"},
		{name: "1.5 MB", bytes: 1024*1024 + 512*1024, expected: "1.50 MB"},
		{name: "exactly 1 GB", bytes: 1024 * 1024 * 1024, expected: "1.00 GB"},
		{name: "1.5 GB", bytes: 1024*1024*1024 + 512*1024*1024, expected: "1.50 GB"},
		{name: "just below 1 MB", bytes: 1024*1024 - 1, expected: "1024.00 KB"},
		{name: "just below 1 GB", bytes: 1024*1024*1024 - 1, expected: "1024.00 MB"},
		{name: "large value", bytes: 5*1024*1024*1024 + 256*1024*1024, expected: "5.25 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSize(tt.bytes)
			if got != tt.expected {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}
