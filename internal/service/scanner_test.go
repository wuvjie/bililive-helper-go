package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsFileBeingWritten_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.flv")
	got := isFileBeingWritten(path, 10*time.Millisecond)
	if !got {
		t.Error("expected true for nonexistent file, got false")
	}
}

func TestIsFileBeingWritten_StableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stable.flv")
	if err := os.WriteFile(path, make([]byte, 4096), 0644); err != nil {
		t.Fatal(err)
	}
	got := isFileBeingWritten(path, 10*time.Millisecond)
	if got {
		t.Error("expected false for stable file, got true")
	}
}

func TestIsFileBeingWritten_SizeChanging(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "growing.flv")
	if err := os.WriteFile(path, make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	// Start a goroutine that appends data to simulate active writing.
	done := make(chan struct{})
	go func() {
		defer close(done)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Millisecond)
			f.Write(make([]byte, 512))
		}
	}()

	got := isFileBeingWritten(path, 50*time.Millisecond)
	<-done
	if !got {
		t.Error("expected true for changing file, got false")
	}
}

func TestIsFileSizeStable_StableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stable.flv")
	if err := os.WriteFile(path, make([]byte, 8192), 0644); err != nil {
		t.Fatal(err)
	}
	got := isFileSizeStable(path, 10*time.Millisecond)
	if !got {
		t.Error("expected true for stable file, got false")
	}
}

func TestIsFileSizeStable_ChangingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "growing.flv")
	if err := os.WriteFile(path, make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	// Start a goroutine that appends data to simulate active writing.
	done := make(chan struct{})
	go func() {
		defer close(done)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Millisecond)
			f.Write(make([]byte, 512))
		}
	}()

	got := isFileSizeStable(path, 50*time.Millisecond)
	<-done
	if got {
		t.Error("expected false for changing file, got true")
	}
}

func TestIsFileSizeStable_NonexistentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.flv")
	got := isFileSizeStable(path, 10*time.Millisecond)
	if got {
		t.Error("expected false for nonexistent file, got true")
	}
}
