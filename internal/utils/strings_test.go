package utils

import (
	"regexp"
	"testing"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		keywords []string
		expected bool
	}{
		{
			name:     "match found",
			s:        "hello world",
			keywords: []string{"world"},
			expected: true,
		},
		{
			name:     "no match",
			s:        "hello world",
			keywords: []string{"xyz"},
			expected: false,
		},
		{
			name:     "case insensitive match",
			s:        "Hello World",
			keywords: []string{"WORLD"},
			expected: true,
		},
		{
			name:     "case insensitive keyword",
			s:        "hello world",
			keywords: []string{"HELLO"},
			expected: true,
		},
		{
			name:     "empty keywords list",
			s:        "hello world",
			keywords: []string{},
			expected: false,
		},
		{
			name:     "nil keywords list",
			s:        "hello world",
			keywords: nil,
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			keywords: []string{"hello"},
			expected: false,
		},
		{
			name:     "multiple keywords first misses second hits",
			s:        "hello world",
			keywords: []string{"xyz", "world"},
			expected: true,
		},
		{
			name:     "multiple keywords all miss",
			s:        "hello world",
			keywords: []string{"xyz", "abc"},
			expected: false,
		},
		{
			name:     "partial substring match",
			s:        "hello world",
			keywords: []string{"ell"},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsAny(tt.s, tt.keywords)
			if got != tt.expected {
				t.Errorf("ContainsAny(%q, %v) = %v, want %v", tt.s, tt.keywords, got, tt.expected)
			}
		})
	}
}

func TestRandomHex(t *testing.T) {
	t.Run("correct length", func(t *testing.T) {
		tests := []int{1, 4, 8, 16, 32}
		for _, n := range tests {
			result := RandomHex(n)
			if len(result) != 2*n {
				t.Errorf("RandomHex(%d): len = %d, want %d", n, len(result), 2*n)
			}
		}
	})

	t.Run("valid hex characters", func(t *testing.T) {
		re := regexp.MustCompile(`^[0-9a-f]+$`)
		for i := 0; i < 10; i++ {
			result := RandomHex(16)
			if !re.MatchString(result) {
				t.Errorf("RandomHex(16) = %q, contains non-hex characters", result)
			}
		}
	})

	t.Run("different calls return different values", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			result := RandomHex(16)
			if seen[result] {
				t.Errorf("RandomHex(16) returned duplicate value %q on iteration %d", result, i)
			}
			seen[result] = true
		}
	})

	t.Run("zero bytes", func(t *testing.T) {
		result := RandomHex(0)
		if result != "" {
			t.Errorf("RandomHex(0) = %q, want empty string", result)
		}
	})
}
