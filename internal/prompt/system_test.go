package prompt

import (
	"strings"
	"testing"
)

func TestSystemPrompt_ContainsResolution(t *testing.T) {
	p := SystemPrompt(1920, 1080, "test task")
	if !strings.Contains(p, "1920x1080") {
		t.Error("system prompt missing resolution 1920x1080")
	}
}

func TestSystemPrompt_ContainsTask(t *testing.T) {
	p := SystemPrompt(800, 600, "open firefox and search for cats")
	if !strings.Contains(p, "open firefox and search for cats") {
		t.Error("system prompt missing task description")
	}
}

func TestSystemPrompt_NotEmpty(t *testing.T) {
	p := SystemPrompt(1920, 1080, "task")
	if p == "" {
		t.Error("system prompt is empty")
	}
}

func TestSystemPrompt_ContainsAllActionTypes(t *testing.T) {
	actionTypes := []string{
		"mouse_move",
		"click",
		"double_click",
		"type_text",
		"key_press",
		"scroll",
		"wait",
		"screenshot",
		"done",
	}

	p := SystemPrompt(1920, 1080, "task")
	for _, at := range actionTypes {
		if !strings.Contains(p, at) {
			t.Errorf("system prompt missing action type %q", at)
		}
	}
}

func TestSystemPrompt_ContainsResponseFormat(t *testing.T) {
	p := SystemPrompt(1920, 1080, "task")
	if !strings.Contains(p, "RESPONSE FORMAT") {
		t.Error("system prompt missing RESPONSE FORMAT section")
	}
}

func TestSystemPrompt_ContainsThinking(t *testing.T) {
	p := SystemPrompt(1920, 1080, "task")
	if !strings.Contains(p, "thinking") {
		t.Error("system prompt missing 'thinking'")
	}
}

func TestSystemPrompt_ContainsActions(t *testing.T) {
	p := SystemPrompt(1920, 1080, "task")
	if !strings.Contains(p, `"actions"`) {
		t.Error("system prompt missing 'actions'")
	}
}

func TestSystemPrompt_DifferentResolutions(t *testing.T) {
	tests := []struct {
		w, h int
		want string
	}{
		{1366, 768, "1366x768"},
		{2560, 1440, "2560x1440"},
		{3840, 2160, "3840x2160"},
	}
	for _, tt := range tests {
		p := SystemPrompt(tt.w, tt.h, "task")
		if !strings.Contains(p, tt.want) {
			t.Errorf("prompt for %dx%d missing %q", tt.w, tt.h, tt.want)
		}
	}
}

func TestSystemPrompt_ContainsCoordinateSystem(t *testing.T) {
	p := SystemPrompt(1920, 1080, "task")
	if !strings.Contains(p, "origin") {
		t.Error("system prompt missing coordinate system info")
	}
}

func TestSystemPrompt_ContainsGuidelines(t *testing.T) {
	p := SystemPrompt(1920, 1080, "task")
	if !strings.Contains(p, "GUIDELINES") {
		t.Error("system prompt missing GUIDELINES section")
	}
}

func TestSystemPrompt_EmptyTask(t *testing.T) {
	p := SystemPrompt(1920, 1080, "")
	if p == "" {
		t.Error("system prompt with empty task should not be empty")
	}
	if !strings.Contains(p, "1920x1080") {
		t.Error("system prompt with empty task missing resolution")
	}
}
