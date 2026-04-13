package action

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type mockInput struct {
	moves      [][2]int
	clicks     []string
	doubleClks int
	texts      []string
	keyCombos  [][]string
	scrolls    []scrollCall
}

type scrollCall struct {
	direction string
	amount    int
}

func (m *mockInput) MoveMouse(x, y int) error {
	m.moves = append(m.moves, [2]int{x, y})
	return nil
}

func (m *mockInput) Click(button string) error {
	m.clicks = append(m.clicks, button)
	return nil
}

func (m *mockInput) DoubleClick() error {
	m.doubleClks++
	return nil
}

func (m *mockInput) TypeText(text string) error {
	m.texts = append(m.texts, text)
	return nil
}

type errInput struct {
	err error
}

func (e *errInput) MoveMouse(int, int) error { return e.err }
func (e *errInput) Click(string) error       { return e.err }
func (e *errInput) DoubleClick() error       { return e.err }
func (e *errInput) TypeText(string) error    { return e.err }
func (e *errInput) PressKey([]string) error  { return e.err }
func (e *errInput) Scroll(string, int) error { return e.err }

func (m *mockInput) PressKey(keys []string) error {
	m.keyCombos = append(m.keyCombos, keys)
	return nil
}

func (m *mockInput) Scroll(direction string, amount int) error {
	m.scrolls = append(m.scrolls, scrollCall{direction, amount})
	return nil
}

func newMockExec() (*Executor, *mockInput) {
	m := &mockInput{}
	return NewExecutor(m, false), m
}

func TestExecuteDone(t *testing.T) {
	e, _ := newMockExec()

	done, err := e.Execute([]Action{{Type: "done"}})
	if err != nil {
		t.Fatalf("Execute done: %v", err)
	}
	if !done {
		t.Error("Execute done = false, want true")
	}
}

func TestExecuteDoneWithReason(t *testing.T) {
	e, _ := newMockExec()

	done, err := e.Execute([]Action{{Type: "done", Reason: "task completed successfully"}})
	if err != nil {
		t.Fatalf("Execute done: %v", err)
	}
	if !done {
		t.Error("Execute done = false, want true")
	}
}

func TestExecuteWait(t *testing.T) {
	e, _ := newMockExec()

	start := time.Now()
	done, err := e.Execute([]Action{{Type: "wait", Duration: 20}})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Execute wait: %v", err)
	}
	if done {
		t.Error("Execute wait = true, want false")
	}
	if elapsed < 15*time.Millisecond {
		t.Errorf("wait elapsed = %v, expected at least ~20ms", elapsed)
	}
}

func TestExecuteScreenshot(t *testing.T) {
	e, _ := newMockExec()

	done, err := e.Execute([]Action{{Type: "screenshot"}})
	if err != nil {
		t.Fatalf("Execute screenshot: %v", err)
	}
	if done {
		t.Error("Execute screenshot = true, want false")
	}
}

func TestExecuteUnknown(t *testing.T) {
	e, _ := newMockExec()

	done, err := e.Execute([]Action{{Type: "teleport"}})
	if err != nil {
		t.Fatalf("Execute unknown: %v", err)
	}
	if done {
		t.Error("Execute unknown = true, want false")
	}
}

func TestExecuteMouseMove(t *testing.T) {
	e, m := newMockExec()

	done, err := e.Execute([]Action{{Type: "mouse_move", X: 100, Y: 200}})
	if err != nil {
		t.Fatalf("Execute mouse_move: %v", err)
	}
	if done {
		t.Error("Execute mouse_move = true, want false")
	}
	if len(m.moves) != 1 {
		t.Fatalf("moves = %d, want 1", len(m.moves))
	}
	if m.moves[0] != [2]int{100, 200} {
		t.Errorf("move = %v, want [100 200]", m.moves[0])
	}
}

func TestExecuteClickDefaultButton(t *testing.T) {
	e, m := newMockExec()

	done, err := e.Execute([]Action{{Type: "click"}})
	if err != nil {
		t.Fatalf("Execute click: %v", err)
	}
	if done {
		t.Error("Execute click = true, want false")
	}
	if len(m.clicks) != 1 || m.clicks[0] != "left" {
		t.Errorf("clicks = %v, want [left]", m.clicks)
	}
}

func TestExecuteClickExplicitButton(t *testing.T) {
	e, m := newMockExec()

	_, err := e.Execute([]Action{{Type: "click", Button: "right"}})
	if err != nil {
		t.Fatalf("Execute click right: %v", err)
	}
	if len(m.clicks) != 1 || m.clicks[0] != "right" {
		t.Errorf("clicks = %v, want [right]", m.clicks)
	}
}

func TestExecuteClickMiddle(t *testing.T) {
	e, m := newMockExec()

	_, err := e.Execute([]Action{{Type: "click", Button: "middle"}})
	if err != nil {
		t.Fatalf("Execute click middle: %v", err)
	}
	if len(m.clicks) != 1 || m.clicks[0] != "middle" {
		t.Errorf("clicks = %v, want [middle]", m.clicks)
	}
}

func TestExecuteDoubleClick(t *testing.T) {
	e, m := newMockExec()

	done, err := e.Execute([]Action{{Type: "double_click"}})
	if err != nil {
		t.Fatalf("Execute double_click: %v", err)
	}
	if done {
		t.Error("Execute double_click = true, want false")
	}
	if m.doubleClks != 1 {
		t.Errorf("doubleClks = %d, want 1", m.doubleClks)
	}
}

func TestExecuteTypeText(t *testing.T) {
	e, m := newMockExec()

	done, err := e.Execute([]Action{{Type: "type_text", Text: "hello world"}})
	if err != nil {
		t.Fatalf("Execute type_text: %v", err)
	}
	if done {
		t.Error("Execute type_text = true, want false")
	}
	if len(m.texts) != 1 || m.texts[0] != "hello world" {
		t.Errorf("texts = %v, want [hello world]", m.texts)
	}
}

func TestExecuteKeyPress(t *testing.T) {
	e, m := newMockExec()

	done, err := e.Execute([]Action{{Type: "key_press", Keys: []string{"ctrl", "c"}}})
	if err != nil {
		t.Fatalf("Execute key_press: %v", err)
	}
	if done {
		t.Error("Execute key_press = true, want false")
	}
	if len(m.keyCombos) != 1 {
		t.Fatalf("keyCombos = %d, want 1", len(m.keyCombos))
	}
	if len(m.keyCombos[0]) != 2 || m.keyCombos[0][0] != "ctrl" || m.keyCombos[0][1] != "c" {
		t.Errorf("keyCombo = %v, want [ctrl c]", m.keyCombos[0])
	}
}

func TestExecuteScrollDefaults(t *testing.T) {
	e, m := newMockExec()

	done, err := e.Execute([]Action{{Type: "scroll"}})
	if err != nil {
		t.Fatalf("Execute scroll: %v", err)
	}
	if done {
		t.Error("Execute scroll = true, want false")
	}
	if len(m.scrolls) != 1 {
		t.Fatalf("scrolls = %d, want 1", len(m.scrolls))
	}
	if m.scrolls[0].direction != "down" {
		t.Errorf("scroll direction = %q, want %q", m.scrolls[0].direction, "down")
	}
	if m.scrolls[0].amount != 3 {
		t.Errorf("scroll amount = %d, want 3", m.scrolls[0].amount)
	}
}

func TestExecuteScrollExplicit(t *testing.T) {
	e, m := newMockExec()

	_, err := e.Execute([]Action{{Type: "scroll", Direction: "up", Amount: 5}})
	if err != nil {
		t.Fatalf("Execute scroll up: %v", err)
	}
	if len(m.scrolls) != 1 {
		t.Fatalf("scrolls = %d, want 1", len(m.scrolls))
	}
	if m.scrolls[0].direction != "up" {
		t.Errorf("scroll direction = %q, want %q", m.scrolls[0].direction, "up")
	}
	if m.scrolls[0].amount != 5 {
		t.Errorf("scroll amount = %d, want 5", m.scrolls[0].amount)
	}
}

func TestExecuteMultipleActions(t *testing.T) {
	e, m := newMockExec()

	actions := []Action{
		{Type: "mouse_move", X: 50, Y: 75},
		{Type: "click", Button: "left"},
		{Type: "type_text", Text: "test"},
		{Type: "done"},
	}

	done, err := e.Execute(actions)
	if err != nil {
		t.Fatalf("Execute multiple: %v", err)
	}
	if !done {
		t.Error("Execute multiple = false, want true")
	}

	if len(m.moves) != 1 || m.moves[0] != [2]int{50, 75} {
		t.Errorf("moves = %v, want [[50 75]]", m.moves)
	}
	if len(m.clicks) != 1 || m.clicks[0] != "left" {
		t.Errorf("clicks = %v, want [left]", m.clicks)
	}
	if len(m.texts) != 1 || m.texts[0] != "test" {
		t.Errorf("texts = %v, want [test]", m.texts)
	}
}

func TestExecuteActionOrder(t *testing.T) {
	e, m := newMockExec()

	actions := []Action{
		{Type: "mouse_move", X: 10, Y: 20},
		{Type: "mouse_move", X: 30, Y: 40},
	}

	_, err := e.Execute(actions)
	if err != nil {
		t.Fatalf("Execute order: %v", err)
	}
	if len(m.moves) != 2 {
		t.Fatalf("moves = %d, want 2", len(m.moves))
	}
	if m.moves[0] != [2]int{10, 20} {
		t.Errorf("first move = %v, want [10 20]", m.moves[0])
	}
	if m.moves[1] != [2]int{30, 40} {
		t.Errorf("second move = %v, want [30 40]", m.moves[1])
	}
}

func TestExecuteStopsOnError(t *testing.T) {
	expectedErr := errInput{err: errors.New("mock input error")}
	e := NewExecutor(&expectedErr, false)

	actions := []Action{
		{Type: "mouse_move", X: 1, Y: 2},
		{Type: "done"},
	}

	done, err := e.Execute(actions)
	if err == nil {
		t.Fatal("Execute should return error when input fails")
	}
	if done {
		t.Error("Execute = true on error, want false")
	}
	if !strings.Contains(err.Error(), "mouse_move") {
		t.Errorf("error = %v, want containing 'mouse_move'", err)
	}
}

func TestExecuteEmptyActions(t *testing.T) {
	e, _ := newMockExec()

	done, err := e.Execute([]Action{})
	if err != nil {
		t.Fatalf("Execute empty: %v", err)
	}
	if done {
		t.Error("Execute empty = true, want false")
	}
}

func TestExecuteNilActions(t *testing.T) {
	e, _ := newMockExec()

	done, err := e.Execute(nil)
	if err != nil {
		t.Fatalf("Execute nil: %v", err)
	}
	if done {
		t.Error("Execute nil = true, want false")
	}
}
