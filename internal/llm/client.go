package llm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"waygent/internal/action"
)

// Client is an OpenAI-compatible API client.
type Client struct {
	apiURL  string
	apiKey  string
	model   string
	verbose bool
}

// NewClient creates a new LLM client.
func NewClient(apiURL, apiKey, model string, verbose bool) *Client {
	return &Client{
		apiURL:  strings.TrimRight(apiURL, "/"),
		apiKey:  apiKey,
		model:   model,
		verbose: verbose,
	}
}

// Message represents a chat message.
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// ContentPart represents a part of a multimodal message.
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL holds the image URL or base64 data.
type ImageURL struct {
	URL string `json:"url"`
}

// ChatRequest is the request body for chat completions.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse is the response body from chat completions.
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Send sends a chat completion request with message history and a new screenshot.
func (c *Client) Send(imagePath string, task string, history []Message) (string, []Message, error) {
	// Read and base64-encode the screenshot
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", nil, fmt.Errorf("reading screenshot %s: %w", imagePath, err)
	}
	b64 := base64.StdEncoding.EncodeToString(imgData)
	dataURI := "data:image/png;base64," + b64

	// Build the user message with text + image
	userContent := []ContentPart{
		{
			Type: "text",
			Text: task,
		},
		{
			Type: "image_url",
			ImageURL: &ImageURL{
				URL: dataURI,
			},
		},
	}

	userMsg := Message{
		Role:    "user",
		Content: userContent,
	}

	// Build full message list: system prompt is expected to be in history[0]
	allMsgs := make([]Message, 0, len(history)+1)
	allMsgs = append(allMsgs, history...)
	allMsgs = append(allMsgs, userMsg)

	reqBody := ChatRequest{
		Model:    c.model,
		Messages: allMsgs,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := c.apiURL + "/chat/completions"
	if c.verbose {
		fmt.Fprintf(os.Stderr, "  LLM request: POST %s (model=%s, messages=%d)\n", url, c.model, len(allMsgs))
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("reading response: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "  LLM response: status %d (%d bytes)\n", resp.StatusCode, len(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", nil, fmt.Errorf("parsing response JSON: %w\nbody: %s", err, string(respBody))
	}

	if chatResp.Error != nil {
		return "", nil, fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from LLM")
	}

	assistantContent := chatResp.Choices[0].Message.Content

	updatedHistory := make([]Message, 0, len(history)+2)
	updatedHistory = append(updatedHistory, history...)
	updatedHistory = append(updatedHistory, userMsg)
	updatedHistory = append(updatedHistory, Message{
		Role:    "assistant",
		Content: assistantContent,
	})

	return assistantContent, updatedHistory, nil
}

// ParseResponse extracts an LLMResponse from raw LLM output content,
// handling markdown code fences and other common wrapping patterns.
func ParseResponse(content string) (*action.LLMResponse, error) {
	cleaned := extractJSON(content)

	var resp action.LLMResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w\ncontent: %s", err, content)
	}
	return &resp, nil
}

// extractJSON strips markdown code fences and extracts the JSON object.
func extractJSON(content string) string {
	s := strings.TrimSpace(content)

	// Try to find content within ```json ... ``` or ``` ... ```
	// Use simple string scanning instead of regex to avoid import.
	for {
		// Look for opening fence
		idx := strings.Index(s, "```")
		if idx == -1 {
			break
		}
		// Find the end of the opening fence line
		afterOpen := s[idx+3:]
		// Skip optional "json" or "JSON" tag
		lineEnd := strings.Index(afterOpen, "\n")
		if lineEnd == -1 {
			break
		}
		jsonStart := afterOpen[lineEnd+1:]
		// Find closing fence
		closeIdx := strings.Index(jsonStart, "```")
		if closeIdx == -1 {
			break
		}
		s = strings.TrimSpace(jsonStart[:closeIdx])
		break
	}

	// If it doesn't start with '{', try to find the outermost JSON object
	if !strings.HasPrefix(s, "{") {
		first := strings.Index(s, "{")
		if first != -1 {
			s = s[first:]
		}
	}

	// Trim trailing content after the last '}'
	last := strings.LastIndex(s, "}")
	if last != -1 {
		s = s[:last+1]
	}

	return s
}
