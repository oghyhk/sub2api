# Sticky Session and Weekly Warm-up Invariants

This document defines the non-negotiable routing behavior for the Google
Antigravity account pool. Any change to session hashing, the digest store,
account selection, or weekly-usage scheduling must preserve these rules.

## Sticky conversation identity

- A continuing Gemini conversation must remain on its assigned account while
  that account is schedulable.
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

## Weekly usage-bar warm-up

- The seven-day warm-up policy remains active for known weekly-use candidates
  below `1%`. It is a scheduling rule, not a reason to discard an established
  sticky conversation.
- Usage at or above `1%` means the account has warmed up and returns to normal
  sticky-session scheduling. Values such as `0.01%` remain in warm-up.
- Changes to sticky-session recovery must not disable, shortcut, or reinterpret
  this policy. The regression test `TestAntigravityWeeklyWarmupOverridesThenRestoresSticky`
  protects this boundary.

## Required verification before deployment

Run the focused tests below and retain a pre-deployment database backup:

```powershell
go test -tags unit ./internal/service -run 'Test(BuildGeminiDigestChain|GenerateGeminiPrefixHash|AntigravityWeeklyWarmup)'
go test -tags unit ./internal/service -run 'TestDigestSession'
```

Then make a multi-turn Gemini request through the public Cloudflare endpoint and
verify the access log reports the same `account_id` across the conversation.
