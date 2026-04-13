package input

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// keyCodes maps key name strings to Linux input-event-codes.h KEY_* values.
var keyCodes = map[string]uint16{
	// Letters
	"a": 30, "b": 48, "c": 46, "d": 32, "e": 18,
	"f": 33, "g": 34, "h": 35, "i": 23, "j": 36,
	"k": 37, "l": 38, "m": 50, "n": 49, "o": 24,
	"p": 25, "q": 16, "r": 19, "s": 31, "t": 20,
	"u": 22, "v": 47, "w": 17, "x": 45, "y": 21,
	"z": 44,
	// Digits
	"0": 11, "1": 2, "2": 3, "3": 4, "4": 5,
	"5": 6, "6": 7, "7": 8, "8": 9, "9": 10,
	// Special keys
	"enter": 28, "return": 28, "escape": 1, "esc": 1,
	"tab": 15, "backspace": 14, "delete": 111, "space": 57,
	"insert": 110, "capslock": 58,
	// Modifiers
	"shift": 42, "leftshift": 42, "rightshift": 54,
	"ctrl": 29, "leftctrl": 29, "rightctrl": 97,
	"alt": 56, "leftalt": 56, "rightalt": 100,
	"meta": 125, "super": 125, "leftmeta": 125,
	"rightmeta": 126, "windows": 125,
	// Function keys
	"f1": 59, "f2": 60, "f3": 61, "f4": 62,
	"f5": 63, "f6": 64, "f7": 65, "f8": 66,
	"f9": 67, "f10": 68, "f11": 87, "f12": 88,
	// Arrow keys
	"up": 103, "down": 108, "left": 105, "right": 106,
	// Navigation keys
	"home": 102, "end": 107, "pageup": 104, "pagedown": 109,
}

// Executor runs input commands via ydotool.
type Executor struct {
	verbose bool
}

// NewExecutor creates a new input executor.
func NewExecutor(verbose bool) *Executor {
	return &Executor{verbose: verbose}
}

func (e *Executor) run(args ...string) error {
	if e.verbose {
		fmt.Fprintf(os.Stderr, "  ydotool %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command("ydotool", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ydotool %s: %w: %s", strings.Join(args, " "), err, string(out))
	}
	return nil
}

// MoveMouse moves the cursor to absolute coordinates (x, y).
func (e *Executor) MoveMouse(x, y int) error {
	return e.run("mousemove", "--absolute", "-x", fmt.Sprintf("%d", x), "-y", fmt.Sprintf("%d", y))
}

// Click performs a mouse click. button: "left", "right", "middle".
func (e *Executor) Click(button string) error {
	var code string
	switch strings.ToLower(button) {
	case "left", "":
		code = "0xC0"
	case "right":
		code = "0xC1"
	case "middle":
		code = "0xC2"
	default:
		return fmt.Errorf("unknown mouse button: %s", button)
	}
	return e.run("click", code)
}

// DoubleClick performs a double left click.
func (e *Executor) DoubleClick() error {
	if err := e.Click("left"); err != nil {
		return fmt.Errorf("double click first: %w", err)
	}
	return e.Click("left")
}

// TypeText types a string using the keyboard.
func (e *Executor) TypeText(text string) error {
	return e.run("type", "--", text)
}

// PressKey presses and releases a key combination.
// For combos: press all keys down in order, then release all in reverse order.
func (e *Executor) PressKey(keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("no keys specified")
	}

	var args []string
	args = append(args, "key")

	// Resolve key codes
	codes := make([]uint16, len(keys))
	for i, k := range keys {
		code, ok := keyCodes[strings.ToLower(k)]
		if !ok {
			return fmt.Errorf("unknown key: %s", k)
		}
		codes[i] = code
	}

	// Press all keys down in order
	for _, code := range codes {
		args = append(args, fmt.Sprintf("%d:1", code))
	}
	// Release all keys in reverse order
	for i := len(codes) - 1; i >= 0; i-- {
		args = append(args, fmt.Sprintf("%d:0", codes[i]))
	}

	return e.run(args...)
}

// Scroll performs a mouse scroll at current position.
func (e *Executor) Scroll(direction string, amount int) error {
	var code string
	switch strings.ToLower(direction) {
	case "up":
		code = "0xC3"
	case "down":
		code = "0xC4"
	default:
		return fmt.Errorf("unknown scroll direction: %s (use up or down)", direction)
	}
	for i := 0; i < amount; i++ {
		if err := e.run("click", code); err != nil {
			return fmt.Errorf("scroll step %d/%d: %w", i+1, amount, err)
		}
	}
	return nil
}
