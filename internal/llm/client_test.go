package llm

import (
	"encoding/json"
	"strings"
	"testing"

	"waygent/internal/action"
)

func TestExtractJSON_PlainJSON(t *testing.T) {
	input := `{"thinking": "hello", "actions": []}`
	got := extractJSON(input)
	if !strings.HasPrefix(got, "{") {
		t.Errorf("extractJSON plain = %q, want starting with '{'", got)
	}
	var resp action.LLMResponse
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Errorf("extractJSON result not valid JSON: %v", err)
	}
}

func TestExtractJSON_JSONCodeFence(t *testing.T) {
	input := "```json\n{\"thinking\": \"test\", \"actions\": []}\n```"
	got := extractJSON(input)
	want := `{"thinking": "test", "actions": []}`
	if strings.TrimSpace(got) != want {
		t.Errorf("extractJSON code fence = %q, want %q", got, want)
	}
}

func TestExtractJSON_PlainCodeFence(t *testing.T) {
	input := "```\n{\"thinking\": \"test\", \"actions\": []}\n```"
	got := extractJSON(input)
	want := `{"thinking": "test", "actions": []}`
	if strings.TrimSpace(got) != want {
		t.Errorf("extractJSON plain fence = %q, want %q", got, want)
	}
}

func TestExtractJSON_TextAround(t *testing.T) {
	input := "Here is my response:\n{\"thinking\": \"test\", \"actions\": []}\nThat's it!"
	got := extractJSON(input)
	if !strings.HasPrefix(got, "{") {
		t.Errorf("extractJSON text around = %q, want starting with '{'", got)
	}
	if !strings.HasSuffix(got, "}") {
		t.Errorf("extractJSON text around = %q, want ending with '}'", got)
	}
}

func TestExtractJSON_Whitespace(t *testing.T) {
	input := "  \n  ```json\n  {\"thinking\": \"test\", \"actions\": []}  \n  ```  \n  "
	got := extractJSON(input)
	var resp action.LLMResponse
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Errorf("extractJSON whitespace result not valid JSON: %v (got %q)", err, got)
	}
}

func TestExtractJSON_NestedJSON(t *testing.T) {
	input := `{"thinking": "test", "actions": [{"type": "click", "button": "left"}]}`
	got := extractJSON(input)
	var resp action.LLMResponse
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Errorf("extractJSON nested result not valid JSON: %v", err)
	}
}

func TestExtractJSON_NoClosingFence(t *testing.T) {
	input := "```json\n{\"thinking\": \"test\""
	got := extractJSON(input)
	if !strings.HasPrefix(got, "{") {
		t.Errorf("extractJSON no closing fence = %q, want starting with '{'", got)
	}
}

func TestExtractJSON_MultipleCodeFences(t *testing.T) {
	input := "some text ```json\n{\"thinking\": \"first\"}\n``` more text"
	got := extractJSON(input)
	want := `{"thinking": "first"}`
	if strings.TrimSpace(got) != want {
		t.Errorf("extractJSON multiple fences = %q, want %q", got, want)
	}
}

func TestParseResponse_ValidJSON(t *testing.T) {
	input := `{"thinking": "hello world", "actions": []}`
	resp, err := ParseResponse(input)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if resp.Thinking != "hello world" {
		t.Errorf("Thinking = %q, want %q", resp.Thinking, "hello world")
	}
	if len(resp.Actions) != 0 {
		t.Errorf("Actions = %d, want 0", len(resp.Actions))
	}
}

func TestParseResponse_WithActions(t *testing.T) {
	input := `{"thinking": "clicking", "actions": [{"type": "click", "button": "left"}, {"type": "mouse_move", "x": 100, "y": 200}]}`
	resp, err := ParseResponse(input)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(resp.Actions) != 2 {
		t.Fatalf("Actions len = %d, want 2", len(resp.Actions))
	}
	if resp.Actions[0].Type != "click" {
		t.Errorf("Action[0].Type = %q, want %q", resp.Actions[0].Type, "click")
	}
	if resp.Actions[0].Button != "left" {
		t.Errorf("Action[0].Button = %q, want %q", resp.Actions[0].Button, "left")
	}
	if resp.Actions[1].Type != "mouse_move" {
		t.Errorf("Action[1].Type = %q, want %q", resp.Actions[1].Type, "mouse_move")
	}
	if resp.Actions[1].X != 100 || resp.Actions[1].Y != 200 {
		t.Errorf("Action[1] coords = (%d,%d), want (100,200)", resp.Actions[1].X, resp.Actions[1].Y)
	}
}

func TestParseResponse_CodeFenceWrapped(t *testing.T) {
	input := "```json\n{\"thinking\": \"fenced\", \"actions\": [{\"type\": \"done\"}]}\n```"
	resp, err := ParseResponse(input)
	if err != nil {
		t.Fatalf("ParseResponse fenced: %v", err)
	}
	if resp.Thinking != "fenced" {
		t.Errorf("Thinking = %q, want %q", resp.Thinking, "fenced")
	}
	if len(resp.Actions) != 1 || resp.Actions[0].Type != "done" {
		t.Errorf("Actions = %v, want [done]", resp.Actions)
	}
}

func TestParseResponse_InvalidJSON(t *testing.T) {
	input := "this is not json at all"
	_, err := ParseResponse(input)
	if err == nil {
		t.Error("ParseResponse invalid JSON: want error, got nil")
	}
}

func TestParseResponse_MissingThinking(t *testing.T) {
	input := `{"actions": [{"type": "click"}]}`
	resp, err := ParseResponse(input)
	if err != nil {
		t.Fatalf("ParseResponse missing thinking: %v", err)
	}
	if resp.Thinking != "" {
		t.Errorf("Thinking = %q, want empty string", resp.Thinking)
	}
}

func TestParseResponse_MissingActions(t *testing.T) {
	input := `{"thinking": "no actions"}`
	resp, err := ParseResponse(input)
	if err != nil {
		t.Fatalf("ParseResponse missing actions: %v", err)
	}
	if len(resp.Actions) != 0 {
		t.Errorf("Actions = %d, want 0", len(resp.Actions))
	}
}

func TestParseResponse_AllActionFields(t *testing.T) {
	input := `{
		"thinking": "full test",
		"actions": [{
			"type": "scroll",
			"x": 10,
			"y": 20,
			"button": "right",
			"text": "hello",
			"keys": ["ctrl", "a"],
			"direction": "up",
			"amount": 5,
			"duration_ms": 100,
			"reason": "done reason"
		}]
	}`
	resp, err := ParseResponse(input)
	if err != nil {
		t.Fatalf("ParseResponse all fields: %v", err)
	}
	a := resp.Actions[0]
	if a.X != 10 || a.Y != 20 {
		t.Errorf("coords = (%d,%d), want (10,20)", a.X, a.Y)
	}
	if a.Button != "right" {
		t.Errorf("Button = %q, want %q", a.Button, "right")
	}
	if a.Text != "hello" {
		t.Errorf("Text = %q, want %q", a.Text, "hello")
	}
	if len(a.Keys) != 2 || a.Keys[0] != "ctrl" || a.Keys[1] != "a" {
		t.Errorf("Keys = %v, want [ctrl a]", a.Keys)
	}
	if a.Direction != "up" {
		t.Errorf("Direction = %q, want %q", a.Direction, "up")
	}
	if a.Amount != 5 {
		t.Errorf("Amount = %d, want 5", a.Amount)
	}
	if a.Duration != 100 {
		t.Errorf("Duration = %d, want 100", a.Duration)
	}
	if a.Reason != "done reason" {
		t.Errorf("Reason = %q, want %q", a.Reason, "done reason")
	}
}
