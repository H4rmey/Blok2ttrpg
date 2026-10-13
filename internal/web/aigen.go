// This file implements the AI character generator endpoints. The browser owns
// the credentials (BYOK) and sends them on each request; the server relays the
// provider call, validates every reply with the engine and drives the retry
// loop. No key or endpoint is ever persisted server-side.
package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/ai"
	"github.com/harmey/blok2ttrpg-v5/internal/docs"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/export"
)

// maxAttempts is the total number of provider calls: the initial attempt plus
// the five retries the feature allows.
const maxAttempts = 6

// aiEvent is one newline-delimited progress record streamed to the browser.
type aiEvent struct {
	Type    string     `json:"type"`
	Message string     `json:"message,omitempty"`
	Attempt int        `json:"attempt,omitempty"`
	Max     int        `json:"max,omitempty"`
	Valid   bool       `json:"valid,omitempty"`
	YAML    string     `json:"yaml,omitempty"`
	Report  *ai.Report `json:"report,omitempty"`
}

// aiReference memoizes the config-derived rules summary handed to the model.
func (a *App) aiReference() string {
	a.aiRefOnce.Do(func() { a.aiRef = docs.AIReference(a.Cfg, a.Library) })
	return a.aiRef
}

// handleAIGenerate runs the generate-validate-retry loop and streams progress
// as newline-delimited JSON. It does not save the character: the browser posts
// the accepted YAML to /ai/import so the keep/discard choice stays client-side.
func (a *App) handleAIGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Description string `json:"description"`
		Endpoint    string `json:"endpoint"`
		APIKey      string `json:"apiKey"`
		Model       string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Description) == "" {
		http.Error(w, "a character description is required", http.StatusBadRequest)
		return
	}
	settings := ai.Settings{Endpoint: req.Endpoint, APIKey: req.APIKey, Model: req.Model}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, _ := w.(http.Flusher)
	emit := func(e aiEvent) {
		b, err := json.Marshal(e)
		if err != nil {
			return
		}
		w.Write(append(b, '\n'))
		if flusher != nil {
			flusher.Flush()
		}
	}

	msgs := []ai.Message{
		{Role: "system", Content: ai.SystemPrompt(a.Cfg, a.aiReference())},
		{Role: "user", Content: userPrompt(req.Description)},
	}

	var lastYAML string
	var lastReport *ai.Report

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		emit(aiEvent{Type: "status", Message: fmt.Sprintf("Generating character... (attempt %d/%d)", attempt, maxAttempts)})

		// Stream the reply so the user can watch it arrive and cancel a slow
		// generation; the browser shows its own rotating status and elapsed time.
		reply, err := ai.ChatStream(r.Context(), settings, msgs, nil)
		if err != nil {
			if r.Context().Err() != nil {
				return // the browser closed or cancelled the request
			}
			emit(aiEvent{Type: "error", Message: err.Error()})
			return
		}

		cleaned := extractYAML(reply)
		// Lead the model: rewrite tier names, passive spellings and the
		// hand-written consumed_effect_ib into the canonical import shape
		// before parsing, so a character that is correct in spirit is not
		// rejected on a cosmetic mismatch.
		yamlDoc := cleaned
		if normalized, nerr := ai.NormalizeYAML(a.Cfg.Config, []byte(cleaned)); nerr == nil {
			yamlDoc = string(normalized)
		}
		c, perr := export.UnmarshalCharacter([]byte(yamlDoc))
		var issue string
		var report ai.Report
		if perr != nil {
			issue = "the reply could not be parsed as character YAML (" + perr.Error() + ")."
		} else {
			report = ai.Validate(a.Cfg.Config, c)
			if report.Valid {
				emit(aiEvent{Type: "done", Valid: true, YAML: yamlDoc, Report: &report})
				return
			}
			lastYAML = yamlDoc
			lastReport = &report
			issue = strings.Join(report.Issues, " ")
		}

		if attempt == maxAttempts {
			break
		}
		emit(aiEvent{
			Type:    "retry",
			Attempt: attempt + 1,
			Max:     maxAttempts,
			Message: fmt.Sprintf("Attempt %d failed: %s Retrying...", attempt, issue),
		})

		msgs = append(msgs, ai.Message{Role: "assistant", Content: reply})
		if perr != nil {
			msgs = append(msgs, ai.Message{
				Role:    "user",
				Content: "Your previous reply could not be parsed as character YAML. Return the full corrected character YAML only, with no markdown fences or commentary.",
			})
		} else {
			msgs = append(msgs, ai.Message{Role: "user", Content: ai.Correction(report)})
		}
	}

	msg := fmt.Sprintf("Failed to generate a valid character after %d attempts.", maxAttempts)
	if lastReport != nil {
		msg = fmt.Sprintf("Failed to generate a valid character after %d attempts: %s", maxAttempts, strings.Join(lastReport.Issues, " "))
	}
	emit(aiEvent{Type: "done", Valid: false, YAML: lastYAML, Report: lastReport, Message: msg})
}

// handleAIGeneratePerk runs the same generate-validate-retry loop as the
// character generator, but for a single perk. It streams progress as
// newline-delimited JSON and does not save: the browser posts the accepted
// perk YAML to /ai/import-perk.
func (a *App) handleAIGeneratePerk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Description string `json:"description"`
		Character   string `json:"character"`
		Endpoint    string `json:"endpoint"`
		APIKey      string `json:"apiKey"`
		Model       string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Description) == "" {
		http.Error(w, "a perk description is required", http.StatusBadRequest)
		return
	}
	settings := ai.Settings{Endpoint: req.Endpoint, APIKey: req.APIKey, Model: req.Model}

	// The character's free perk points are advisory guidance for the model, not
	// a hard budget: perk cost is advisory in this app, so an over-budget perk
	// can still be generated and kept.
	remaining := 0
	if req.Character != "" {
		if c, ok := a.Store.Get(req.Character); ok {
			stats := a.characterStats(&c)
			remaining = stats.PerkBudget - stats.PerkUsed
		}
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, _ := w.(http.Flusher)
	emit := func(e aiEvent) {
		b, err := json.Marshal(e)
		if err != nil {
			return
		}
		w.Write(append(b, '\n'))
		if flusher != nil {
			flusher.Flush()
		}
	}

	msgs := []ai.Message{
		{Role: "system", Content: ai.PerkSystemPrompt(a.Cfg, a.aiReference())},
		{Role: "user", Content: ai.PerkUserPrompt(req.Description, remaining)},
	}

	var lastYAML string
	var lastIssues []string

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		emit(aiEvent{Type: "status", Message: fmt.Sprintf("Generating perk... (attempt %d/%d)", attempt, maxAttempts)})

		reply, err := ai.ChatStream(r.Context(), settings, msgs, nil)
		if err != nil {
			if r.Context().Err() != nil {
				return
			}
			emit(aiEvent{Type: "error", Message: err.Error()})
			return
		}

		yamlDoc := extractYAML(reply)
		p, perr := export.UnmarshalPerk([]byte(yamlDoc))
		var issue string
		if perr != nil {
			issue = "the reply could not be parsed as perk YAML (" + perr.Error() + ")."
		} else {
			p = engine.NormalizePerk(a.Cfg.Config, p)
			issues := ai.ValidatePerk(a.Cfg.Config, p)
			if len(issues) == 0 {
				out, merr := export.MarshalPerk(p)
				if merr != nil {
					emit(aiEvent{Type: "error", Message: merr.Error()})
					return
				}
				emit(aiEvent{Type: "done", Valid: true, YAML: string(out)})
				return
			}
			lastYAML = yamlDoc
			lastIssues = issues
			issue = strings.Join(issues, " ")
		}

		if attempt == maxAttempts {
			break
		}
		emit(aiEvent{
			Type:    "retry",
			Attempt: attempt + 1,
			Max:     maxAttempts,
			Message: fmt.Sprintf("Attempt %d failed: %s Retrying...", attempt, issue),
		})

		msgs = append(msgs, ai.Message{Role: "assistant", Content: reply})
		if perr != nil {
			msgs = append(msgs, ai.Message{
				Role:    "user",
				Content: "Your previous reply could not be parsed as perk YAML. Return the full corrected perk YAML only, with no markdown fences or commentary.",
			})
		} else {
			msgs = append(msgs, ai.Message{
				Role: "user",
				Content: "Your perk failed validation. " + issue +
					" A non-passive perk needs at least one enactment; use the passive shape (type: passive with fields.passive_id) instead if that fits." +
					" Return the full corrected perk YAML only, with no markdown fences or commentary.",
			})
		}
	}

	msg := fmt.Sprintf("Failed to generate a valid perk after %d attempts.", maxAttempts)
	if len(lastIssues) > 0 {
		msg = fmt.Sprintf("Failed to generate a valid perk after %d attempts: %s", maxAttempts, strings.Join(lastIssues, " "))
	}
	emit(aiEvent{Type: "done", Valid: false, YAML: lastYAML, Message: msg})
}

// handleAIImportPerk parses a generated perk YAML and appends it to the
// character named by the "character" query parameter.
func (a *App) handleAIImportPerk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	charID := r.URL.Query().Get("character")
	c, ok := a.Store.Get(charID)
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unknown character")
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "could not read body")
		return
	}
	p, err := export.UnmarshalPerk([]byte(extractYAML(string(data))))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	p.ID = fmt.Sprintf("perk-%d", time.Now().UnixNano())
	p = engine.NormalizePerk(a.Cfg.Config, p)
	if reason := perkRequiresEnactment(a.Cfg.Config, p); reason != "" {
		writeJSONError(w, http.StatusBadRequest, reason)
		return
	}
	if ok, reason := a.affordPerk(&c, p); !ok {
		writeJSONError(w, http.StatusBadRequest, reason)
		return
	}
	c.Perks = append(c.Perks, p)
	if err := a.Store.Save(c); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": p.ID})
}

// handleAIModels relays GET {endpoint}/models so the model dropdown works on
// providers that send no CORS headers.
func (a *App) handleAIModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	var req struct {
		Endpoint string `json:"endpoint"`
		APIKey   string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	models, err := ai.ListModels(r.Context(), req.Endpoint, req.APIKey)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"models": models})
}

// handleAIImport parses a generated character YAML and stores it through the
// same repair path a hand-written import uses.
func (a *App) handleAIImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "could not read body")
		return
	}
	c, err := export.UnmarshalCharacter([]byte(extractYAML(string(data))))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := a.saveImportedCharacter(c)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

// userPrompt frames the user's description. The level is the model's choice,
// so the prompt asks it to pick one that fits and spend that level's budgets.
func userPrompt(desc string) string {
	return fmt.Sprintf("Character description:\n%s\n\nChoose a level that fits the description, then generate the character YAML now, spending exactly the full skill and perk budgets for that level.", strings.TrimSpace(desc))
}

// extractYAML strips a surrounding markdown code fence if the model added one,
// so a reply that is otherwise correct still parses.
func extractYAML(s string) string {
	t := strings.TrimSpace(s)
	i := strings.Index(t, "```")
	if i < 0 {
		return t
	}
	rest := t[i+3:]
	// Drop an optional language tag on the opening fence line.
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		first := strings.TrimSpace(rest[:nl])
		if first == "" || !strings.Contains(first, ":") {
			rest = rest[nl+1:]
		}
	}
	if j := strings.LastIndex(rest, "```"); j >= 0 {
		rest = rest[:j]
	}
	return strings.TrimSpace(rest)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg})
}
