package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestChatStreamAccumulatesDeltas pins the SSE path: streamed content fragments
// are concatenated and reported through onDelta.
func TestChatStreamAccumulatesDeltas(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		for _, part := range []string{"level: ", "1\n", "skills:\n"} {
			chunk, _ := json.Marshal(map[string]any{
				"choices": []map[string]any{{"delta": map[string]any{"content": part}}},
			})
			w.Write([]byte("data: " + string(chunk) + "\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	var deltas []string
	got, err := ChatStream(context.Background(), Settings{Endpoint: srv.URL, APIKey: "k", Model: "m"},
		[]Message{{Role: "user", Content: "hi"}}, func(d string) { deltas = append(deltas, d) })
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}
	if got != "level: 1\nskills:\n" {
		t.Errorf("accumulated = %q", got)
	}
	if len(deltas) != 3 {
		t.Errorf("onDelta called %d times, want 3", len(deltas))
	}
	if gotBody["stream"] != true {
		t.Errorf("request did not ask for streaming: %v", gotBody["stream"])
	}
	if _, ok := gotBody["max_tokens"]; !ok {
		t.Error("request did not set max_tokens")
	}
}

// TestChatStreamFallsBackToPlainCompletion covers providers that ignore the
// stream flag and return an ordinary completion.
func TestChatStreamFallsBackToPlainCompletion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "plain reply"}}},
		})
	}))
	defer srv.Close()

	got, err := ChatStream(context.Background(), Settings{Endpoint: srv.URL, APIKey: "k", Model: "m"}, nil, nil)
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}
	if strings.TrimSpace(got) != "plain reply" {
		t.Errorf("got %q", got)
	}
}
