package capture

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
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
// Tries methods in order: XDG Desktop Portal, GNOME Shell D-Bus,
// gnome-screenshot, grim.
func Capture(path string) error {
	// Method 1: XDG Desktop Portal (works on GNOME Wayland 42+)
	if err := capturePortal(path); err == nil {
		return nil
	}
	// Method 2: GNOME Shell Screenshot via D-Bus (older GNOME)
	if err := captureGNOMEShell(path); err == nil {
		return nil
	}
	// Method 3: gnome-screenshot
	if err := captureGnomeScreenshot(path); err == nil {
		return nil
	}
	// Method 4: grim (wlroots-based compositors)
	if err := captureGrim(path); err == nil {
		return nil
	}
	return fmt.Errorf("all screenshot methods failed (portal, gdbus, gnome-screenshot, grim)")
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

// capturePortal uses the XDG Desktop Portal Screenshot API.
// This is the standard way to capture screenshots on GNOME Wayland 42+,
// where direct D-Bus and grim access are blocked by security policies.
// The portal call is async: Screenshot() returns a Request handle, and the
// actual file URI arrives via the Response signal on that handle.
func capturePortal(targetPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	monitorCmd := exec.CommandContext(ctx, "gdbus", "monitor", "--session",
		"--dest", "org.freedesktop.portal.Desktop")
	stdoutPipe, err := monitorCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("portal monitor: %w", err)
	}
	if err := monitorCmd.Start(); err != nil {
		return fmt.Errorf("portal monitor start: %w", err)
	}
	defer monitorCmd.Process.Kill()

	time.Sleep(200 * time.Millisecond)

	callOut, err := exec.CommandContext(ctx, "gdbus", "call", "--session",
		"--dest", "org.freedesktop.portal.Desktop",
		"--object-path", "/org/freedesktop/portal/desktop",
		"--method", "org.freedesktop.portal.Screenshot.Screenshot",
		"{}", "{}").Output()
	if err != nil {
		return fmt.Errorf("portal call: %w: %s", err, string(callOut))
	}

	scanner := bufio.NewScanner(stdoutPipe)

	for scanner.Scan() {
		line := scanner.Text()
		code, uri, matched, perr := parsePortalResponseLine(line)
		if !matched {
			continue
		}
		if perr != nil {
			return fmt.Errorf("portal response code %d: %w", code, perr)
		}
		return copyFile(uri, targetPath)
	}

	return fmt.Errorf("portal: no response received within timeout")
}

var portalResponseRe = regexp.MustCompile(`org\.freedesktop\.portal\.Request\.Response\((\d+),`)
var portalURIRe = regexp.MustCompile(`'uri':\s*<\s*'([^']+)'`)

// parsePortalResponseLine extracts the response code and file URI from a
// gdbus monitor line containing a portal Response signal.
// Returns (code, uri, matched, error). matched is false if the line is not
// a Response signal. error is set for non-zero response codes or missing URI.
func parsePortalResponseLine(line string) (code int, uri string, matched bool, err error) {
	if !strings.Contains(line, "org.freedesktop.portal.Request.Response") {
		return 0, "", false, nil
	}
	matched = true

	codeMatches := portalResponseRe.FindStringSubmatch(line)
	if len(codeMatches) < 2 {
		return 0, "", true, fmt.Errorf("could not parse response code")
	}
	code, _ = strconv.Atoi(codeMatches[1])
	if code != 0 {
		return code, "", true, fmt.Errorf("screenshot request denied or failed")
	}

	uriMatches := portalURIRe.FindStringSubmatch(line)
	if len(uriMatches) < 2 {
		return 0, "", true, fmt.Errorf("no URI in response")
	}
	uri = strings.TrimPrefix(uriMatches[1], "file://")
	return 0, uri, true, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy source: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copy dest: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}
	return out.Close()
}
