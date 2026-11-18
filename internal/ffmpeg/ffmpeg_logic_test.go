package ffmpeg

import (
	"context"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestVideoBSFForCodec(t *testing.T) {
	tests := []struct {
		codec string
		want  []string
	}{
		{codec: "h264", want: []string{"-bsf:v", "h264_mp4toannexb"}},
		{codec: "AVC1", want: []string{"-bsf:v", "h264_mp4toannexb"}},
		{codec: "hevc", want: []string{"-bsf:v", "hevc_mp4toannexb"}},
		{codec: "H265", want: []string{"-bsf:v", "hevc_mp4toannexb"}},
		{codec: "av1", want: nil},
		{codec: "", want: nil},
	}
	for _, tt := range tests {
		got := VideoBSFForCodec(tt.codec)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("VideoBSFForCodec(%q) = %v, want %v", tt.codec, got, tt.want)
		}
	}
}

func TestBuildConcatFilter(t *testing.T) {
	t.Run("all have audio", func(t *testing.T) {
		filter, useAudio := buildConcatFilter(2, []bool{true, true}, []float64{0, 0})
		want := "[0:v:0][0:a:0][1:v:0][1:a:0]concat=n=2:v=1:a=1[outv][outa]"
		if !useAudio || filter != want {
			t.Fatalf("got (%q, %v), want (%q, true)", filter, useAudio, want)
		}
	})

	t.Run("none have audio", func(t *testing.T) {
		filter, useAudio := buildConcatFilter(2, []bool{false, false}, []float64{0, 0})
		want := "[0:v:0][1:v:0]concat=n=2:v=1:a=0[outv]"
		if useAudio || filter != want {
			t.Fatalf("got (%q, %v), want (%q, false)", filter, useAudio, want)
		}
	})

	t.Run("mixed audio with known durations pads silence", func(t *testing.T) {
		filter, useAudio := buildConcatFilter(2, []bool{true, false}, []float64{0, 12.5})
		want := "anullsrc=r=44100:cl=stereo:d=12.500[silent_1];[0:v:0][0:a:0][1:v:0][silent_1]concat=n=2:v=1:a=1[outv][outa]"
		if !useAudio || filter != want {
			t.Fatalf("got (%q, %v), want (%q, true)", filter, useAudio, want)
		}
	})

	t.Run("mixed audio with missing duration falls back to video only", func(t *testing.T) {
		filter, useAudio := buildConcatFilter(2, []bool{true, false}, []float64{0, 0})
		want := "[0:v:0][1:v:0]concat=n=2:v=1:a=0[outv]"
		if useAudio || filter != want {
			t.Fatalf("got (%q, %v), want (%q, false)", filter, useAudio, want)
		}
	})

	t.Run("cross resolution adds scale pad and setsar filters", func(t *testing.T) {
		filter, useAudio := buildConcatFilterWithScale(2, []bool{true, true}, []float64{0, 0}, []int{1920, 1280}, []int{1080, 720})
		want := "[0:v:0]scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2,setsar=1[v_0];" +
			"[1:v:0]scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2,setsar=1[v_1];" +
			"[v_0][0:a:0][v_1][1:a:0]concat=n=2:v=1:a=1[outv][outa]"
		if !useAudio || filter != want {
			t.Fatalf("got (%q, %v), want (%q, true)", filter, useAudio, want)
		}
	})
}

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg binary not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe binary not installed")
	}
}

func makeSampleVideo(t *testing.T, path, format, codec string, withAudio bool, durationSec string) {
	t.Helper()
	args := []string{
		"-y", "-v", "error",
		"-f", "lavfi", "-i", "testsrc=size=160x120:rate=25:duration=" + durationSec,
	}
	if withAudio {
		args = append(args, "-f", "lavfi", "-i", "sine=frequency=440:duration="+durationSec)
	}
	args = append(args, "-c:v", codec)
	if codec == "libx264" {
		args = append(args, "-preset", "ultrafast", "-pix_fmt", "yuv420p")
	}
	if withAudio {
		args = append(args, "-c:a", "aac", "-shortest")
	}
	if format != "" {
		args = append(args, "-f", format)
	}
	args = append(args, path)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := Run(ctx, Options{Args: args}); err != nil {
		t.Fatalf("generate sample video %s failed: %v", path, err)
	}
}

func TestIntegration_ConvertTrimProbeAndMixedReencode(t *testing.T) {
	requireFFmpeg(t)
	ctx := context.Background()
	dir := t.TempDir()

	// 1. FLV H.264 -> MP4 direct convert + codec probe + TrimVideo + IsVideoTrimmed
	flvPath := filepath.Join(dir, "sample.flv")
	mp4Path := filepath.Join(dir, "sample.mp4")
	makeSampleVideo(t, flvPath, "flv", "libx264", true, "3")

	if codec := ProbeVideoCodec(ctx, flvPath); codec != "h264" {
		t.Fatalf("ProbeVideoCodec(%s) = %q, want h264", flvPath, codec)
	}
	if err := ConvertViaTS(ctx, flvPath, mp4Path); err != nil {
		t.Fatalf("ConvertViaTS failed: %v", err)
	}
	if err := ValidateOutput(ctx, mp4Path); err != nil {
		t.Fatalf("ValidateOutput(converted) failed: %v", err)
	}
	if IsVideoTrimmed(ctx, mp4Path) {
		t.Fatal("expected IsVideoTrimmed=false before trimming")
	}

	trimmedPath := filepath.Join(dir, "trimmed.mp4")
	if err := TrimVideo(ctx, mp4Path, trimmedPath, "00:00:01", "00:00:03"); err != nil {
		t.Fatalf("TrimVideo failed: %v", err)
	}
	if !IsVideoTrimmed(ctx, trimmedPath) {
		t.Fatal("expected IsVideoTrimmed=true after trimming")
	}

	// 2. Single TS -> MP4 convert
	tsPath := filepath.Join(dir, "single.ts")
	tsOut := filepath.Join(dir, "single.mp4")
	makeSampleVideo(t, tsPath, "mpegts", "libx264", true, "2")
	if err := ConvertViaTS(ctx, tsPath, tsOut); err != nil {
		t.Fatalf("ConvertViaTS(.ts -> .mp4) failed: %v", err)
	}
	if err := ValidateOutput(ctx, tsOut); err != nil {
		t.Fatalf("ValidateOutput(ts->mp4) failed: %v", err)
	}

	// 3. Reencode mixed audio / silent segments
	segWithAudio := "seg1.mp4"
	segSilent := "seg2.mp4"
	makeSampleVideo(t, filepath.Join(dir, segWithAudio), "mp4", "libx264", true, "2")
	makeSampleVideo(t, filepath.Join(dir, segSilent), "mp4", "libx264", false, "2")
	mixedOut := filepath.Join(dir, "mixed_merged.mp4")
	if err := Reencode(ctx, []string{segWithAudio, segSilent}, dir, mixedOut, nil); err != nil {
		t.Fatalf("Reencode(mixed audio/silent) failed: %v", err)
	}
	if !HasAudioStream(ctx, mixedOut) {
		t.Fatal("expected merged mixed-audio video to contain audio stream")
	}
	dur, err := ProbeDuration(ctx, mixedOut)
	if err != nil || dur < 3.0 {
		t.Fatalf("merged mixed-audio duration = %.2f (err=%v), want >= 3.0s", dur, err)
	}
}
