package capture

import (
	"image"
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
