package action

import (
	"fmt"
	"os"
	"time"
)

// Action represents a single desktop action returned by the LLM.
type Action struct {
	Type      string   `json:"type"`
	X         int      `json:"x,omitempty"`
	Y         int      `json:"y,omitempty"`
	Button    string   `json:"button,omitempty"`
	Text      string   `json:"text,omitempty"`
	Keys      []string `json:"keys,omitempty"`
	Direction string   `json:"direction,omitempty"`
	Amount    int      `json:"amount,omitempty"`
	Duration  int      `json:"duration_ms,omitempty"`
	Reason    string   `json:"reason,omitempty"`
}

// LLMResponse is the structured response expected from the LLM.
type LLMResponse struct {
	Thinking string   `json:"thinking"`
	Actions  []Action `json:"actions"`
}

// InputExecutor is the interface for desktop input operations.
type InputExecutor interface {
	MoveMouse(x, y int) error
	Click(button string) error
	DoubleClick() error
	TypeText(text string) error
	PressKey(keys []string) error
	Scroll(direction string, amount int) error
}

// Executor runs a slice of actions using the input executor.
type Executor struct {
	input   InputExecutor
	verbose bool
}

// NewExecutor creates an action executor.
func NewExecutor(inputExec InputExecutor, verbose bool) *Executor {
	return &Executor{input: inputExec, verbose: verbose}
}

// Execute runs all actions in sequence. Returns (done bool, err error).
func (e *Executor) Execute(actions []Action) (bool, error) {
	for _, a := range actions {
		if e.verbose {
			fmt.Fprintf(os.Stderr, "  action: %+v\n", a)
		}

		switch a.Type {
		case "done":
			reason := a.Reason
			if reason == "" {
				reason = "task complete"
			}
			fmt.Fprintf(os.Stderr, "done: %s\n", reason)
			return true, nil

		case "wait":
			dur := a.Duration
			if dur <= 0 {
				dur = 500
			}
			if e.verbose {
				fmt.Fprintf(os.Stderr, "  waiting %dms\n", dur)
			}
			time.Sleep(time.Duration(dur) * time.Millisecond)

		case "screenshot":
			// No-op signal; the agent loop handles the actual capture.

		case "mouse_move":
			if err := e.input.MoveMouse(a.X, a.Y); err != nil {
				return false, fmt.Errorf("mouse_move to (%d,%d): %w", a.X, a.Y, err)
			}

		case "click":
			button := a.Button
			if button == "" {
				button = "left"
			}
			if err := e.input.Click(button); err != nil {
				return false, fmt.Errorf("click %s: %w", button, err)
			}

		case "double_click":
			if err := e.input.DoubleClick(); err != nil {
				return false, fmt.Errorf("double_click: %w", err)
			}

		case "type_text":
			if err := e.input.TypeText(a.Text); err != nil {
				return false, fmt.Errorf("type_text: %w", err)
			}

		case "key_press":
			if err := e.input.PressKey(a.Keys); err != nil {
				return false, fmt.Errorf("key_press %v: %w", a.Keys, err)
			}

		case "scroll":
			amount := a.Amount
			if amount <= 0 {
				amount = 3
			}
			direction := a.Direction
			if direction == "" {
				direction = "down"
			}
			if err := e.input.Scroll(direction, amount); err != nil {
				return false, fmt.Errorf("scroll %s %d: %w", direction, amount, err)
			}

		default:
			fmt.Fprintf(os.Stderr, "warning: unknown action type: %s\n", a.Type)
		}
	}
	return false, nil
}
