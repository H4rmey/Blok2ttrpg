// Package ai contains the OpenAI-compatible client and the character
// validation used by the AI character generator. It is deliberately separate
// from internal/web so the provider protocol, the system prompt and the point
// validation can be read and tested without the HTTP plumbing.
//
// All credentials are supplied per request by the user (BYOK): nothing here is
// hardcoded, and no key or endpoint is ever persisted by the server.
package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// maxOutputTokens bounds a single completion, so a model that tries to pad the
// budget with a very long character cannot generate for an unbounded time.
const maxOutputTokens = 8192

// Settings is one user-supplied provider configuration. It mirrors what the
// browser keeps in localStorage and sends on each generation request.
type Settings struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
}

// Message is one entry in the OpenAI chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// httpClient is the shared client. A per-attempt timeout is enforced with the
// request context, not the client, so a slow local model is not cut off by the
// transport while still bounding the whole call. Streaming means the user can
// watch progress and cancel, so the ceiling can be generous.
var httpClient = &http.Client{Timeout: 10 * time.Minute}

// validate checks the shared requirements for a provider call.
func (s Settings) validate() error {
	if strings.TrimSpace(s.Endpoint) == "" {
		return fmt.Errorf("no API endpoint configured")
	}
	if strings.TrimSpace(s.APIKey) == "" {
		return fmt.Errorf("no API key configured")
	}
	if strings.TrimSpace(s.Model) == "" {
		return fmt.Errorf("no model selected")
	}
	return nil
}

// baseURL normalises the user's endpoint and rejects anything that is not an
// http(s) URL. The endpoint is user-supplied (BYOK), so this is a guardrail
// against typos rather than a security boundary.
func (s Settings) baseURL() (string, error) {
	raw := strings.TrimSpace(s.Endpoint)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid endpoint %q", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("endpoint must be http or https")
	}
	return strings.TrimRight(raw, "/"), nil
}

func (s Settings) chatURL() (string, error) {
	base, err := s.baseURL()
	if err != nil {
		return "", err
	}
	return base + "/chat/completions", nil
}

// chatRequest builds a POST to {endpoint}/chat/completions. When stream is
// true the provider is asked to stream tokens as server-sent events.
func (s Settings) chatRequest(ctx context.Context, msgs []Message, stream bool) (*http.Request, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	u, err := s.chatURL()
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"model":       s.Model,
		"messages":    msgs,
		"temperature": 0.7,
		"max_tokens":  maxOutputTokens,
	}
	if stream {
		payload["stream"] = true
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

// Chat sends the conversation and returns the assistant's reply without
// streaming. It is a fallback for providers that ignore the stream flag.
func Chat(ctx context.Context, s Settings, msgs []Message) (string, error) {
	req, err := s.chatRequest(ctx, msgs, false)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling provider: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("reading provider response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("provider returned %d: %s", resp.StatusCode, truncate(string(data), 500))
	}
	return parseCompletion(data)
}

// ChatStream sends the conversation with streaming enabled and calls onDelta
// for every content fragment as it arrives. It returns the full reply text.
//
// Streaming is what makes a slow model tolerable: the caller can show live
// progress and the user can cancel. If the provider ignores the stream flag and
// returns a normal completion, the raw body is parsed as one instead.
func ChatStream(ctx context.Context, s Settings, msgs []Message, onDelta func(string)) (string, error) {
	req, err := s.chatRequest(ctx, msgs, true)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		return "", fmt.Errorf("provider returned %d: %s", resp.StatusCode, truncate(string(data), 500))
	}

	br := bufio.NewReaderSize(resp.Body, 1<<16)
	var acc strings.Builder
	var raw bytes.Buffer
	for {
		line, readErr := br.ReadString('\n')
		if len(line) > 0 {
			raw.WriteString(line)
			if payload, ok := sseData(line); ok {
				if payload == "[DONE]" {
					break
				}
				if delta := completionDelta(payload); delta != "" {
					acc.WriteString(delta)
					if onDelta != nil {
						onDelta(delta)
					}
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			if acc.Len() > 0 {
				break
			}
			return "", fmt.Errorf("reading provider stream: %w", readErr)
		}
	}
	if acc.Len() > 0 {
		return acc.String(), nil
	}
	// No SSE frames: the provider returned an ordinary completion.
	return parseCompletion(raw.Bytes())
}

// sseData extracts the payload of a "data: ..." server-sent-event line.
func sseData(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "data:") {
		return "", false
	}
	return strings.TrimSpace(t[len("data:"):]), true
}

// completionDelta pulls the incremental content out of one streamed chunk.
func completionDelta(payload string) string {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil || len(chunk.Choices) == 0 {
		return ""
	}
	return chunk.Choices[0].Delta.Content
}

// parseCompletion decodes a non-streamed OpenAI chat completion.
func parseCompletion(data []byte) (string, error) {
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("decoding provider response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("provider returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}

// ListModels fetches the available model ids from {endpoint}/models.
func ListModels(ctx context.Context, endpoint, apiKey string) ([]string, error) {
	s := Settings{Endpoint: endpoint, APIKey: apiKey}
	base, err := s.baseURL()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("no API key configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing models: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("reading models response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider returned %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decoding models response: %w", err)
	}
	ids := make([]string, 0, len(out.Data))
	for _, m := range out.Data {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return nil, fmt.Errorf("no models returned by the provider")
	}
	return ids, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
