package priming

import "context"

type TemplateSource struct {
	templates []string
	index     int
}

func NewTemplateSource() *TemplateSource {
	return &TemplateSource{templates: []string{
		"Take three slow breaths. Tell yourself: 'I remember my dreams clearly when I wake.'",
		"Imagine noticing something unusual in a dream, then calmly saying: 'This is a dream.'",
		"Before sleep, visualize waking up and writing one vivid dream detail in your journal.",
	}}
}

func (s *TemplateSource) Label() SourceLabel {
	return SourceTemplate
}

func (s *TemplateSource) Next(ctx context.Context) (string, error) {
	_ = ctx
	if len(s.templates) == 0 {
		return "", errSourceUnavailable
	}

	content := s.templates[s.index%len(s.templates)]
	s.index++
	return content, nil
}
