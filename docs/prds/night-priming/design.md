# Night Priming Design

**Status:** Brainstormed  
**Date:** 2025-03-17  
**Participants:** User, AI

## Problem Statement
Users need help priming their brains at night to increase chances of dream recall and lucidity in the morning. This requires pre-sleep content that reinforces dream awareness.

## Goals
- Prime the brain with dream-related content before sleep
- Use personal dream signs from clustering for relevance
- Provide variety through community content (Reddit) and AI generation
- Work gracefully when network/APIs fail (cascading fallbacks)
- Integrate seamlessly with existing dream journal TUI

## Approach

TUI-based night mode displaying rotating priming content with cascading fallbacks:
1. **Personalized stories** using user's dream signs from clustering
2. **Reddit posts** from r/LucidDreaming (fetched on-demand)
3. **AI-generated stories** via OpenAI-compatible API (fallback)
4. **Static templates** (last resort)

### Alternatives Considered

- **Option B (Hybrid with periodic sync):** Rejected in favor of on-demand fetching to reduce complexity and avoid background processes
- **Option C (Local LLM):** Rejected due to model download complexity; using API approach instead

## Architecture

### New Components

```
internal/priming/
├── generator.go      # Content orchestration logic
├── reddit.go         # Reddit API client (raw HTTP)
├── ai.go             # OpenAI-compatible API client
└── templates.go      # Static fallback templates
```

### TUI Integration

- New view: `nightView` - fullscreen priming display
- Entry key: `p` (from list view)
- Navigation: `n` next content, `esc` exit, optional timer

### Storage

- `priming_cache` table: Cached Reddit posts (24h TTL)
- `priming_logs` table: Track what content was shown when

### Environment Variables

Following pattern from property analyzer:
```
AI_BASE_URL         # e.g., https://api.openai.com/v1
AI_API_KEY          # API key
AI_MODEL            # Primary model (e.g., gpt-4o-mini)
AI_MODEL_FALLBACK   # Optional fallback model
```

## Data Flow

```
User presses 'p' → nightView
→ Generator picks content:

1. Personalized (priority 1)
   - Get latest clusters
   - Extract top dream signs
   - Interpolate into template
   - Display [Personalized]

2. Reddit (priority 2, on-demand)
   - HTTP GET r/LucidDreaming/.json
   - Filter posts (exclude memes, require selftext)
   - Cache to priming_cache
   - Display [Community]

3. AI Generation (fallback)
   - POST to AI_BASE_URL/chat/completions
   - Prompt includes dream signs
   - Display [AI Generated]

4. Static Templates (last resort)
   - Interpolate dream signs
   - Display [Template]
```

## Error Handling

| Failure Mode | Response |
|--------------|----------|
| No clusters available | Skip to Reddit/AI/Static |
| Reddit fetch fails | Skip to AI/Static |
| AI API fails | Skip to Static |
| All sources fail | Show error, allow retry |
| Rate limited | Back off, use lower priority source |

## Testing Strategy

- Unit tests for each generator source
- Mock Reddit API responses
- Mock AI API responses
- Test fallback cascade
- Test env var configuration

## Open Questions

- Should priming content be logged per user? (Yes, for pattern analysis)
- Auto-exit timer duration? (5-15 min configurable)
- Should we track which content sources improve recall? (Future enhancement)
