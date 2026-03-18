package priming

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type SourceLabel string

const (
	SourcePersonalized SourceLabel = "Personalized"
	SourceCommunity    SourceLabel = "Community"
	SourceAI           SourceLabel = "AI Generated"
	SourceTemplate     SourceLabel = "Template"

	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
)

var errSourceUnavailable = fmt.Errorf("source unavailable")

type Source interface {
	Label() SourceLabel
	Next(ctx context.Context) (string, error)
}

type LogStore interface {
	SavePrimingLog(ctx context.Context, source, outcome, detail, content string, createdAt time.Time) error
}

type Result struct {
	Source SourceLabel
	Text   string
	Status string
	Err    error
}

type Generator struct {
	store   LogStore
	sources []Source
	nowFn   func() time.Time
}

func NewGenerator(store LogStore, sources ...Source) *Generator {
	return &Generator{store: store, sources: sources, nowFn: func() time.Time { return time.Now().UTC() }}
}

func (g *Generator) Next(ctx context.Context) Result {
	if len(g.sources) == 0 {
		return Result{Err: fmt.Errorf("no priming sources configured")}
	}

	var failures []string
	for i, source := range g.sources {
		text, err := source.Next(ctx)
		if err == nil {
			g.logAttempt(source.Label(), OutcomeSuccess, "")
			status := degradedStatus(failures, source.Label())
			g.persistOutcome(ctx, source.Label(), OutcomeSuccess, status, text)
			return Result{Source: source.Label(), Text: text, Status: status}
		}

		detail := err.Error()
		g.logAttempt(source.Label(), OutcomeFailure, detail)
		failures = append(failures, fmt.Sprintf("%s: %s", source.Label(), detail))

		if i < len(g.sources)-1 {
			g.logTransition(source.Label(), g.sources[i+1].Label(), detail)
		}
	}

	status := "All priming sources failed. Press n to retry."
	err := errors.New(status)
	g.logTerminalFailure(failures)
	g.persistOutcome(ctx, SourceTemplate, OutcomeFailure, status, "")
	return Result{Source: SourceTemplate, Status: status, Err: err}
}

func degradedStatus(failures []string, recoveredSource SourceLabel) string {
	if len(failures) == 0 {
		return ""
	}

	if hasAIConfigFailure(failures) {
		return fmt.Sprintf("Using %s fallback. Fix AI config: set AI_BASE_URL, AI_API_KEY, and AI_MODEL.", recoveredSource)
	}

	return fmt.Sprintf("Using %s fallback after upstream source errors.", recoveredSource)
}

func hasAIConfigFailure(failures []string) bool {
	for _, failure := range failures {
		if !strings.HasPrefix(failure, string(SourceAI)+":") {
			continue
		}

		if strings.Contains(failure, "AI_BASE_URL") ||
			strings.Contains(failure, "AI_API_KEY") ||
			strings.Contains(failure, "AI_MODEL") {
			return true
		}
	}

	return false
}

func (g *Generator) persistOutcome(ctx context.Context, source SourceLabel, outcome, detail, content string) {
	if g.store == nil {
		return
	}

	err := g.store.SavePrimingLog(ctx, string(source), outcome, detail, content, g.nowFn())
	if err != nil {
		log.Printf("{\"event\":\"priming_log_persist\",\"status\":\"failure\",\"error\":%q}", err.Error())
	}
}

func (g *Generator) logAttempt(source SourceLabel, outcome, detail string) {
	log.Printf("{\"event\":\"priming_source_attempt\",\"source\":%q,\"outcome\":%q,\"detail\":%q}", source, outcome, detail)
}

func (g *Generator) logTransition(from, to SourceLabel, detail string) {
	log.Printf("{\"event\":\"priming_fallback_transition\",\"from\":%q,\"to\":%q,\"reason\":%q}", from, to, detail)
}

func (g *Generator) logTerminalFailure(failures []string) {
	log.Printf("{\"event\":\"priming_terminal_failure\",\"failures\":%q}", failures)
}
