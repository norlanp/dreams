package priming

import (
	"context"
	"sync/atomic"
)

type TemplateSource struct {
	templates []string
	index     atomic.Int64
}

func NewTemplateSource() *TemplateSource {
	ts := &TemplateSource{templates: []string{
		"Take three slow breaths. Tell yourself: 'I remember my dreams clearly when I wake.'",
		"Imagine noticing something unusual in a dream, then calmly saying: 'This is a dream.'",
		"Before sleep, visualize waking up and writing one vivid dream detail in your journal.",
	}}
	ts.index.Store(0)
	return ts
}

func (s *TemplateSource) Label() SourceLabel {
	return SourceTemplate
}

func (s *TemplateSource) Next(ctx context.Context) (string, error) {
	_ = ctx
	if len(s.templates) == 0 {
		return "", errSourceUnavailable
	}

	idx := int(s.index.Add(1)-1) % len(s.templates)
	content := s.templates[idx]
	return content, nil
}
