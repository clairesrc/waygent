package capture

import (
	"fmt"
	"image"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ScreenResolution returns the current screen width and height in pixels.
// Uses xrandr via os/exec to query display geometry, falls back to wlr-randr.
func ScreenResolution() (image.Point, error) {
	// Try xrandr first
	if p, err := screenResolutionXrandr(); err == nil {
		return p, nil
	}
	// Fallback to wlr-randr
	if p, err := screenResolutionWlrRandr(); err == nil {
		return p, nil
	}
	return image.Point{}, fmt.Errorf("failed to determine screen resolution: both xrandr and wlr-randr failed")
}

func screenResolutionXrandr() (image.Point, error) {
	out, err := exec.Command("xrandr", "--current").Output()
	if err != nil {
		return image.Point{}, fmt.Errorf("xrandr: %w", err)
	}
	return parseXRandrOutput(string(out))
}

// parseXRandrOutput finds the first connected display's geometry.
// Format: "   1920x1080+0+0" in the line after a connected output.
func parseXRandrOutput(output string) (image.Point, error) {
	re := regexp.MustCompile(`(\d+)x(\d+)\+\d+\+\d+`)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, " connected") {
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 3 {
				w, err := strconv.Atoi(matches[1])
				if err != nil {
					continue
				}
				h, err := strconv.Atoi(matches[2])
				if err != nil {
					continue
				}
				return image.Pt(w, h), nil
			}
		}
	}
	return image.Point{}, fmt.Errorf("no connected display found in xrandr output")
}

func screenResolutionWlrRandr() (image.Point, error) {
	out, err := exec.Command("wlr-randr").Output()
	if err != nil {
		return image.Point{}, fmt.Errorf("wlr-randr: %w", err)
	}
	return parseWlrRandrOutput(string(out))
}

func parseWlrRandrOutput(output string) (image.Point, error) {
	re := regexp.MustCompile(`(\d+)x(\d+)`)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "current") {
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 3 {
				w, err := strconv.Atoi(matches[1])
				if err != nil {
					continue
				}
				h, err := strconv.Atoi(matches[2])
				if err != nil {
					continue
				}
				return image.Pt(w, h), nil
			}
		}
	}
	return image.Point{}, fmt.Errorf("no display geometry found in wlr-randr output")
}

// Capture takes a full screenshot and saves to the given path as PNG.
// Tries methods in order: GNOME Shell D-Bus, gnome-screenshot, grim.
func Capture(path string) error {
	// Method 1: GNOME Shell Screenshot via D-Bus
	if err := captureGNOMEShell(path); err == nil {
		return nil
	}
	// Method 2: gnome-screenshot
	if err := captureGnomeScreenshot(path); err == nil {
		return nil
	}
	// Method 3: grim (for non-GNOME Wayland compositors)
	if err := captureGrim(path); err == nil {
		return nil
	}
	return fmt.Errorf("all screenshot methods failed (gdbus, gnome-screenshot, grim)")
}

func captureGNOMEShell(path string) error {
	cmd := exec.Command("gdbus", "call", "--session",
		"--dest", "org.gnome.Shell",
		"--object-path", "/org/gnome/Shell/Screenshot",
		"--method", "org.gnome.Shell.Screenshot.Screenshot",
		"false", "false", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gdbus screenshot: %w: %s", err, string(out))
	}
	return nil
}

func captureGnomeScreenshot(path string) error {
	cmd := exec.Command("gnome-screenshot", "-f", path, "-p")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gnome-screenshot: %w: %s", err, string(out))
	}
	return nil
}

func captureGrim(path string) error {
	cmd := exec.Command("grim", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("grim: %w: %s", err, string(out))
	}
	return nil
}
