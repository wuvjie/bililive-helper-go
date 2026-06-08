package utils

import "testing"

func TestMakeOutputName_NormalFlv(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal raw flv file",
			input:    "[2026-06-04 00-19-00][user1][title1]001.flv",
			expected: "[2026-06-04 00-19-00][user1][title1]-合并版.flv",
		},
		{
			name:     "normal raw mp4 file",
			input:    "[2026-06-04 00-19-00][user1][title1]001.mp4",
			expected: "[2026-06-04 00-19-00][user1][title1]-合并版.mp4",
		},
		{
			name:     "ts extension",
			input:    "[2026-06-04 00-19-00][user1][title1]001.ts",
			expected: "[2026-06-04 00-19-00][user1][title1]-合并版.ts",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeOutputName(tt.input)
			if got != tt.expected {
				t.Errorf("MakeOutputName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMakeOutputName_MergedInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "merged file input should not produce double suffix",
			input:    "[2026-06-03 21-06-44][user1][title1]-合并版.mp4",
			expected: "[2026-06-03 21-06-44][user1][title1]-合并版.mp4",
		},
		{
			name:     "merged flv input",
			input:    "[2026-06-03 21-06-44][user1][title1]-合并版.flv",
			expected: "[2026-06-03 21-06-44][user1][title1]-合并版.flv",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeOutputName(tt.input)
			if got != tt.expected {
				t.Errorf("MakeOutputName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMakeOutputName_NoBracket(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no closing bracket falls back to stem + suffix",
			input:    "randomvideo.flv",
			expected: "randomvideo-合并版.flv",
		},
		{
			name:     "plain mp4 no bracket",
			input:    "somefile.mp4",
			expected: "somefile-合并版.mp4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeOutputName(tt.input)
			if got != tt.expected {
				t.Errorf("MakeOutputName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMakeOutputName_ExtensionPreserved(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "flv extension preserved",
			input:    "[2026-06-04 00-19-00][user1][title1]001.flv",
			expected: "[2026-06-04 00-19-00][user1][title1]-合并版.flv",
		},
		{
			name:     "mp4 extension preserved",
			input:    "[2026-06-04 00-19-00][user1][title1]001.mp4",
			expected: "[2026-06-04 00-19-00][user1][title1]-合并版.mp4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeOutputName(tt.input)
			if got != tt.expected {
				t.Errorf("MakeOutputName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMakeOutputName_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple closing brackets uses last",
			input:    "[2026-06-04 00-19-00][streamer][title]extra001.flv",
			expected: "[2026-06-04 00-19-00][streamer][title]-合并版.flv",
		},
		{
			name:     "empty extension",
			input:    "[2026-06-04 00-19-00][user1][title1]001",
			expected: "[2026-06-04 00-19-00][user1][title1]-合并版",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeOutputName(tt.input)
			if got != tt.expected {
				t.Errorf("MakeOutputName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "normal flv file", input: "video001.flv", expected: true},
		{name: "normal mp4 file", input: "record.mp4", expected: true},
		{name: "chinese filename", input: "直播回放.mp4", expected: true},
		{name: "long filename", input: "a]b" + string(make([]byte, 200)) + ".flv", expected: false}, // contains null bytes
		{name: "empty string", input: "", expected: false},
		{name: "dot", input: ".", expected: false},
		{name: "double dot", input: "..", expected: false},
		{name: "dotdot prefix", input: "..video.flv", expected: false},
		{name: "dotdot middle", input: "path..video.flv", expected: false},
		{name: "forward slash", input: "path/video.flv", expected: false},
		{name: "backslash", input: "path\\video.flv", expected: false},
		{name: "null byte", input: "video\x00.flv", expected: false},
		{name: "pipe character", input: "video|test.flv", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateFilename(tt.input)
			if got != tt.expected {
				t.Errorf("ValidateFilename(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateFilename_LongValid(t *testing.T) {
	// A long but valid filename (200+ characters, no special chars)
	long := ""
	for i := 0; i < 100; i++ {
		long += "ab"
	}
	long += ".flv"
	if !ValidateFilename(long) {
		t.Errorf("ValidateFilename(long valid name) = false, want true")
	}
}

func TestIsMergedFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "contains merged tag", input: "[2026-06-04][user1]-合并版.flv", expected: true},
		{name: "merged tag in middle", input: "abc-合并版xyz.flv", expected: true},
		{name: "no merged tag", input: "video001.flv", expected: false},
		{name: "empty string", input: "", expected: false},
		{name: "partial match", input: "合并版.flv", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMergedFile(tt.input)
			if got != tt.expected {
				t.Errorf("IsMergedFile(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "mp4", input: "video.mp4", expected: true},
		{name: "flv", input: "video.flv", expected: true},
		{name: "ts", input: "video.ts", expected: true},
		{name: "uppercase MP4", input: "video.MP4", expected: true},
		{name: "uppercase FLV", input: "video.FLV", expected: true},
		{name: "mkv not supported", input: "video.mkv", expected: false},
		{name: "avi not supported", input: "video.avi", expected: false},
		{name: "no extension", input: "video", expected: false},
		{name: "empty string", input: "", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsVideoFile(tt.input)
			if got != tt.expected {
				t.Errorf("IsVideoFile(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMakeMP4Name(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "flv to mp4", input: "video.flv", expected: "video.mp4"},
		{name: "already mp4", input: "video.mp4", expected: "video.mp4"},
		{name: "ts to mp4", input: "video.ts", expected: "video.mp4"},
		{name: "no extension adds mp4", input: "video", expected: "video.mp4"},
		{name: "path preserved", input: "[2026-06-04 00-19-00][user1][title1]001.flv", expected: "[2026-06-04 00-19-00][user1][title1]001.mp4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeMP4Name(tt.input)
			if got != tt.expected {
				t.Errorf("MakeMP4Name(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
