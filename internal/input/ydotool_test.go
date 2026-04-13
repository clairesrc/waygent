package input

import (
	"strings"
	"testing"
)

func TestKeyCodes_Letters(t *testing.T) {
	spotChecks := map[string]uint16{
		"a": 30, "b": 48, "c": 46, "d": 32, "e": 18,
		"f": 33, "g": 34, "h": 35, "i": 23, "j": 36,
		"k": 37, "l": 38, "m": 50, "n": 49, "o": 24,
		"p": 25, "q": 16, "r": 19, "s": 31, "t": 20,
		"u": 22, "v": 47, "w": 17, "x": 45, "y": 21,
		"z": 44,
	}
	for name, want := range spotChecks {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}

	for c := 'a'; c <= 'z'; c++ {
		name := string(c)
		if _, ok := keyCodes[name]; !ok {
			t.Errorf("missing letter key: %q", name)
		}
	}
}

func TestKeyCodes_Digits(t *testing.T) {
	spotChecks := map[string]uint16{
		"0": 11, "1": 2, "2": 3, "3": 4, "4": 5,
		"5": 6, "6": 7, "7": 8, "8": 9, "9": 10,
	}
	for name, want := range spotChecks {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestKeyCodes_SpecialKeys(t *testing.T) {
	expected := map[string]uint16{
		"enter":     28,
		"escape":    1,
		"tab":       15,
		"space":     57,
		"backspace": 14,
		"delete":    111,
		"insert":    110,
		"capslock":  58,
	}
	for name, want := range expected {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestKeyCodes_Modifiers(t *testing.T) {
	expected := map[string]uint16{
		"shift":      42,
		"leftshift":  42,
		"rightshift": 54,
		"ctrl":       29,
		"leftctrl":   29,
		"rightctrl":  97,
		"alt":        56,
		"leftalt":    56,
		"rightalt":   100,
		"meta":       125,
		"leftmeta":   125,
		"rightmeta":  126,
	}
	for name, want := range expected {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestKeyCodes_Aliases(t *testing.T) {
	aliases := map[string]uint16{
		"return":  28,
		"esc":     1,
		"super":   125,
		"windows": 125,
	}
	for name, want := range aliases {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}

	if keyCodes["return"] != keyCodes["enter"] {
		t.Errorf("return(%d) != enter(%d)", keyCodes["return"], keyCodes["enter"])
	}
	if keyCodes["esc"] != keyCodes["escape"] {
		t.Errorf("esc(%d) != escape(%d)", keyCodes["esc"], keyCodes["escape"])
	}
	if keyCodes["super"] != keyCodes["meta"] {
		t.Errorf("super(%d) != meta(%d)", keyCodes["super"], keyCodes["meta"])
	}
	if keyCodes["windows"] != keyCodes["meta"] {
		t.Errorf("windows(%d) != meta(%d)", keyCodes["windows"], keyCodes["meta"])
	}
}

func TestKeyCodes_FunctionKeys(t *testing.T) {
	expected := map[string]uint16{
		"f1": 59, "f2": 60, "f3": 61, "f4": 62,
		"f5": 63, "f6": 64, "f7": 65, "f8": 66,
		"f9": 67, "f10": 68, "f11": 87, "f12": 88,
	}
	for name, want := range expected {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestKeyCodes_ArrowKeys(t *testing.T) {
	expected := map[string]uint16{
		"up":    103,
		"down":  108,
		"left":  105,
		"right": 106,
	}
	for name, want := range expected {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestKeyCodes_NavigationKeys(t *testing.T) {
	expected := map[string]uint16{
		"home":     102,
		"end":      107,
		"pageup":   104,
		"pagedown": 109,
	}
	for name, want := range expected {
		got, ok := keyCodes[name]
		if !ok {
			t.Errorf("keyCodes[%q] not found", name)
		} else if got != want {
			t.Errorf("keyCodes[%q] = %d, want %d", name, got, want)
		}
	}
}

func TestKeyCodes_UnknownKey(t *testing.T) {
	_, ok := keyCodes["nonexistent"]
	if ok {
		t.Error("keyCodes[\"nonexistent\"] should not exist")
	}
}

func TestPressKey_UnknownKey(t *testing.T) {
	e := NewExecutor(false)
	err := e.PressKey([]string{"ctrl", "nonexistent_key_xyz"})
	if err == nil {
		t.Error("PressKey with unknown key should return error")
	}
	if !strings.Contains(err.Error(), "unknown key") {
		t.Errorf("PressKey error = %v, want error containing 'unknown key'", err)
	}
}

func TestPressKey_EmptyKeys(t *testing.T) {
	e := NewExecutor(false)
	err := e.PressKey([]string{})
	if err == nil {
		t.Error("PressKey with empty keys should return error")
	}
	if !strings.Contains(err.Error(), "no keys specified") {
		t.Errorf("PressKey error = %v, want error containing 'no keys specified'", err)
	}
}

func TestPressKey_NilKeys(t *testing.T) {
	e := NewExecutor(false)
	err := e.PressKey(nil)
	if err == nil {
		t.Error("PressKey with nil keys should return error")
	}
	if !strings.Contains(err.Error(), "no keys specified") {
		t.Errorf("PressKey error = %v, want error containing 'no keys specified'", err)
	}
}

func TestClick_UnknownButton(t *testing.T) {
	e := NewExecutor(false)
	err := e.Click("pinkie")
	if err == nil {
		t.Error("Click with unknown button should return error")
	}
	if !strings.Contains(err.Error(), "unknown mouse button") {
		t.Errorf("Click error = %v, want error containing 'unknown mouse button'", err)
	}
}

func TestScroll_UnknownDirection(t *testing.T) {
	e := NewExecutor(false)
	err := e.Scroll("sideways", 3)
	if err == nil {
		t.Error("Scroll with unknown direction should return error")
	}
	if !strings.Contains(err.Error(), "unknown scroll direction") {
		t.Errorf("Scroll error = %v, want error containing 'unknown scroll direction'", err)
	}
}
