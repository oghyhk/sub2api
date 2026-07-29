# Sticky Session and Weekly Warm-up Invariants

This document defines the non-negotiable routing behavior for every gateway
client (WorkBuddy, OpenCode, Hermes, and native Gemini) that can use the Google
Antigravity account pool. Any change to session hashing, the digest store,
request transforms, account selection, or weekly-usage scheduling must preserve
these rules.

## Sticky conversation identity

- A continuing conversation must remain on its assigned account while that
  account is schedulable. The sliding affinity window is 24 hours.
- An explicit caller session ID is preferred and is API-key scoped. Without one,
  the fallback identity uses the system prompt plus the first user turn, so a
  client that resends growing history does not rotate accounts on every turn.
- The primary session cache is the fast path. The digest store is the recovery
  path when a client changes its request hash, the cache expires, or a proxy
  changes network metadata.
- Digest identity includes actual conversation content (system prompt, message
  content, tools, tool responses, images, and thought markers). It excludes
  `thoughtSignature`, because that value is issued by a specific upstream Google
  account and changes after safe recovery/failover.
- Digest namespaces must not include observed client IPs. Cloudflare and other
  reverse proxies legitimately rotate their egress IP between consecutive turns.
- A missing binding with an old thought signature must clean that signature once,
  establish a new binding, and save the updated digest for future turns.
- Native Gemini forwarding must preserve explicit `sessionId` values. When one
  is absent, it injects a deterministic account/model-scoped value derived from
  the logical session; random values are forbidden.
- Tool-call repair must be deterministic for the same request. Randomly minted
  function-call IDs mutate the historical prefix and cause an avoidable Google
  cache miss even if the account did not rotate.

## Failover boundary

- A short upstream retry delay is retried on the same bound account to preserve
  cache warmth. A delay of **15 seconds or more**, an unavailable account, or a
  hard upstream failure may trigger normal failover.
- Failover is an availability recovery path, not a load-balancing mechanism for
  an otherwise healthy sticky conversation.

## Weekly usage-bar warm-up

- The seven-day warm-up policy remains active for known weekly-use candidates
  below `1%`. It is a scheduling rule, not a reason to discard an established
  sticky conversation.
- Usage at or above `1%` means the account has warmed up and returns to normal
  sticky-session scheduling. Values such as `0.01%` remain in warm-up.
- When a new account must be selected, the scheduler keeps its normal priority
  and load-safety gates, then prefers the account with lower known weekly usage
  before using last-used time as the tie-breaker. An established sticky session
  still wins while its account remains schedulable.
- Changes to sticky-session recovery must not disable, shortcut, or reinterpret
  this policy. The regression test `TestAntigravityWeeklyWarmupOverridesThenRestoresSticky`
  protects this boundary.

## Required verification before deployment

Run the focused tests below and retain a pre-deployment database backup:

```powershell
go test -tags=unit -p 1 ./internal/service -run 'Test(GenerateSessionHash|EnsureGeminiFunctionCallIDs|CachePrefix|EnsureGeminiUpstreamSession|ScopedClient|AntigravityWeeklyWarmup|HandleSmartRetry_ExactlyAtThreshold)'
go test -tags=unit -p 1 ./internal/service -run 'TestDigestSession'
```

Then make multi-turn requests through the OpenAI-compatible and native Gemini
endpoints and verify the cache-affinity log lines retain the same account and
prefix fingerprint across the conversation. Do not log request content or raw
session IDs.
