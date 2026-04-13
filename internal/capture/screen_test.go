package capture

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseXRandrOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    image.Point
		wantErr bool
	}{
		{
			name:   "single connected display",
			output: "DP-1 connected primary 1920x1080+0+0 (normal left inverted right x-axis y-axis) 527mm x 296mm",
			want:   image.Pt(1920, 1080),
		},
		{
			name: "multiple outputs first connected wins",
			output: strings.Join([]string{
				"HDMI-1 disconnected",
				"DP-1 connected primary 2560x1440+0+0 (normal) 597mm x 336mm",
				"DP-2 connected 1920x1080+2560+0 (normal) 527mm x 296mm",
			}, "\n"),
			want: image.Pt(2560, 1440),
		},
		{
			name: "disconnected displays skipped",
			output: strings.Join([]string{
				"HDMI-1 disconnected",
				"DP-2 disconnected",
			}, "\n"),
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
		{
			name:    "no connected display",
			output:  "Screen 0: minimum 320 x 200, current 0 x 0, maximum 16384 x 16384",
			wantErr: true,
		},
		{
			name:    "connected but no geometry on that line geometry on same line",
			output:  "DP-1 connected",
			wantErr: true,
		},
		{
			name: "geometry without connected keyword ignored",
			output: strings.Join([]string{
				"DP-1 disconnected 1920x1080+0+0",
				"HDMI-1 disconnected",
			}, "\n"),
			wantErr: true,
		},
		{
			name:   "1366x768 resolution",
			output: "LVDS-1 connected primary 1366x768+0+0 (normal) 344mm x 194mm",
			want:   image.Pt(1366, 768),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseXRandrOutput(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseXRandrOutput() = %v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseXRandrOutput() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseXRandrOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseWlrRandrOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    image.Point
		wantErr bool
	}{
		{
			name: "current mode found",
			output: strings.Join([]string{
				"eDP-1",
				"  Physical size: 310x170 mm",
				"  current mode: 1920x1080 @ 60 Hz",
				"  Available modes:",
				"    1920x1080 @ 60 Hz",
			}, "\n"),
			want: image.Pt(1920, 1080),
		},
		{
			name: "different resolution",
			output: strings.Join([]string{
				"DP-1",
				"  current mode: 2560x1440 @ 144 Hz",
			}, "\n"),
			want: image.Pt(2560, 1440),
		},
		{
			name: "no matching line",
			output: strings.Join([]string{
				"DP-1",
				"  Physical size: 527x296 mm",
			}, "\n"),
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
		{
			name: "current without resolution",
			output: strings.Join([]string{
				"DP-1",
				"  current: enabled",
			}, "\n"),
			wantErr: true,
		},
		{
			name: "3840x2160 resolution",
			output: strings.Join([]string{
				"DP-1",
				"  current mode: 3840x2160 @ 60 Hz",
			}, "\n"),
			want: image.Pt(3840, 2160),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWlrRandrOutput(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseWlrRandrOutput() = %v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseWlrRandrOutput() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseWlrRandrOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	t.Run("copies content verbatim", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		content := []byte("hello waygent\nline2\n")
		if err := os.WriteFile(src, content, 0644); err != nil {
			t.Fatal(err)
		}
		if err := copyFile(src, dst); err != nil {
			t.Fatalf("copyFile() error: %v", err)
		}
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(content) {
			t.Errorf("copyFile() = %q, want %q", got, content)
		}
	})

	t.Run("source does not exist", func(t *testing.T) {
		dir := t.TempDir()
		dst := filepath.Join(dir, "dst.txt")
		err := copyFile(filepath.Join(dir, "nonexistent"), dst)
		if err == nil {
			t.Error("copyFile() = nil, want error")
		}
	})

	t.Run("destination directory does not exist", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
		err := copyFile(src, filepath.Join(dir, "nodir", "dst.txt"))
		if err == nil {
			t.Error("copyFile() = nil, want error")
		}
	})

	t.Run("binary content", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.bin")
		dst := filepath.Join(dir, "dst.bin")
		content := make([]byte, 4096)
		for i := range content {
			content[i] = byte(i % 256)
		}
		if err := os.WriteFile(src, content, 0644); err != nil {
			t.Fatal(err)
		}
		if err := copyFile(src, dst); err != nil {
			t.Fatalf("copyFile() error: %v", err)
		}
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(content) {
			t.Errorf("copyFile() len = %d, want %d", len(got), len(content))
		}
	})
}

func TestParsePortalResponseLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantCode    int
		wantURI     string
		wantErr     bool
		wantNoMatch bool
	}{
		{
			name:     "successful response with uri",
			line:     `/org/freedesktop/portal/desktop/request/1_42/waygent: org.freedesktop.portal.Request.Response(0, {'uri': <'file:///run/user/1000/doc/abc/Screenshot.png'>})`,
			wantCode: 0,
			wantURI:  "/run/user/1000/doc/abc/Screenshot.png",
		},
		{
			name:     "cancelled response",
			line:     `/org/freedesktop/portal/desktop/request/1_42/waygent: org.freedesktop.portal.Request.Response(1, {})`,
			wantCode: 1,
			wantErr:  true,
		},
		{
			name:     "error response code 2",
			line:     `/org/freedesktop/portal/desktop/request/1_42/waygent: org.freedesktop.portal.Request.Response(2, {})`,
			wantCode: 2,
			wantErr:  true,
		},
		{
			name:        "not a response line",
			line:        `/org/freedesktop/portal/desktop: some other signal`,
			wantNoMatch: true,
		},
		{
			name:     "success code but no uri",
			line:     `/org/freedesktop/portal/desktop/request/1_42/waygent: org.freedesktop.portal.Request.Response(0, {})`,
			wantCode: 0,
			wantErr:  true,
		},
		{
			name:     "uri with spaces in gdbus format",
			line:     `/org/freedesktop/portal/desktop/request/1_42/t: org.freedesktop.portal.Request.Response(0, {'uri': <'file:///tmp/shot.png'>})`,
			wantCode: 0,
			wantURI:  "/tmp/shot.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, uri, matched, err := parsePortalResponseLine(tt.line)
			if tt.wantNoMatch {
				if matched {
					t.Error("parsePortalResponseLine() matched, want no match")
				}
				return
			}
			if !matched {
				t.Fatal("parsePortalResponseLine() did not match Response line")
			}
			if code != tt.wantCode {
				t.Errorf("code = %d, want %d", code, tt.wantCode)
			}
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if uri != tt.wantURI {
				t.Errorf("uri = %q, want %q", uri, tt.wantURI)
			}
		})
	}
}

func TestCaptureErrorMessageIncludesPortal(t *testing.T) {
	err := Capture("/nonexistent/path/waygent-test-fail.png")
	if err == nil {
		t.Fatal("Capture() should fail without screenshot tools")
	}
	if !strings.Contains(err.Error(), "portal") {
		t.Errorf("Capture() error = %q, want error mentioning 'portal'", err.Error())
	}
}
