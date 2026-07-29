# Cache Affinity Guardrails

These rules apply to every change that can affect gateway request identity,
account selection, Antigravity forwarding, or retry/failover behavior.

1. Treat an established conversation binding as stronger than balancing and
   weekly-usage preferences. Rotate only when the bound account is unavailable,
   disabled, or the upstream has returned a retry delay of at least 15 seconds.
2. Keep the affinity window at 24 hours unless an explicitly approved product
   policy replaces it. This applies to OpenAI-compatible, Anthropic-compatible,
   and native Gemini entry points.
3. Request transformations before an upstream call must be deterministic.
   Never mint random IDs in prior conversation history: that changes Google's
   cache prefix and defeats otherwise-correct account stickiness.
4. Prefer an explicit, API-key-scoped client session ID. When a client does not
   supply one, derive the fallback identity from the system prompt and first
   user turn, never from the entire growing history or client IP.
5. Preserve explicit upstream `sessionId` values. For native Gemini requests
   without one, derive the injected value deterministically from session,
   account, and model.
6. Keep the seven-day warm-up rule: accounts below 1% weekly usage are eligible
   warm-up candidates; for a new conversation prefer lower known weekly usage.
   It must never evict a healthy sticky conversation.

Before merging a related change, run the cache-affinity tests named in
`docs/STICKY_SESSION_INVARIANT.md` and update that document if the intended
policy changes.
