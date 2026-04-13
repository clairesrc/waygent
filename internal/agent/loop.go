package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"waygent/internal/action"
	"waygent/internal/capture"
	"waygent/internal/config"
	"waygent/internal/input"
	"waygent/internal/llm"
	"waygent/internal/prompt"
)

// Agent runs the main agent loop.
type Agent struct {
	config     *config.Config
	client     *llm.Client
	actionExec *action.Executor
}

// New creates a new Agent.
func New(cfg *config.Config) *Agent {
	inputExec := input.NewExecutor(cfg.Verbose)
	return &Agent{
		config:     cfg,
		client:     llm.NewClient(cfg.APIURL, cfg.APIKey, cfg.Model, cfg.Verbose),
		actionExec: action.NewExecutor(inputExec, cfg.Verbose),
	}
}

// Run executes the agent loop. Returns error on failure.
func (a *Agent) Run() error {
	// 1. Get screen resolution
	res, err := capture.ScreenResolution()
	if err != nil {
		return fmt.Errorf("getting screen resolution: %w", err)
	}
	fmt.Fprintf(os.Stderr, "screen resolution: %dx%d\n", res.X, res.Y)

	// 2. Initialize message history with system prompt
	systemPrompt := prompt.SystemPrompt(res.X, res.Y, a.config.Task)
	history := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 3. Agent loop
	for step := 1; step <= a.config.MaxSteps; step++ {
		fmt.Fprintf(os.Stderr, "\n[step %d/%d]\n", step, a.config.MaxSteps)

		// a. Ensure temp dir exists
		if err := os.MkdirAll(a.config.TempDir, 0o755); err != nil {
			return fmt.Errorf("creating temp dir %s: %w", a.config.TempDir, err)
		}

		// b. Generate screenshot path
		screenshotPath := filepath.Join(a.config.TempDir, fmt.Sprintf("waygent-step-%04d.png", step))

		// c. Capture screenshot
		if a.config.Verbose {
			fmt.Fprintf(os.Stderr, "  capturing screenshot: %s\n", screenshotPath)
		}
		if err := capture.Capture(screenshotPath); err != nil {
			return fmt.Errorf("capturing screenshot: %w", err)
		}

		// d. Send to LLM
		response, updatedHistory, err := a.client.Send(screenshotPath, a.config.Task, history)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: LLM request failed: %v\n", err)
			// Brief sleep and retry
			time.Sleep(1 * time.Second)
			continue
		}
		history = updatedHistory

		if a.config.Verbose {
			fmt.Fprintf(os.Stderr, "  LLM response (%d chars):\n%s\n", len(response), truncate(response, 500))
		}

		// e. Parse response JSON
		llmResp, err := llm.ParseResponse(response)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: failed to parse LLM response: %v\n", err)
			// Retry — the LLM might return malformed JSON
			time.Sleep(500 * time.Millisecond)
			continue
		}

		fmt.Fprintf(os.Stderr, "  thinking: %s\n", llmResp.Thinking)

		// g. Execute actions
		done, err := a.actionExec.Execute(llmResp.Actions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: action execution error: %v\n", err)
			// Continue loop — try to recover
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// h. If done, return
		if done {
			return nil
		}

		// j. Brief sleep between iterations
		time.Sleep(200 * time.Millisecond)
	}

	// 4. Max steps reached
	fmt.Fprintf(os.Stderr, "warning: reached maximum steps (%d) without completion\n", a.config.MaxSteps)
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
