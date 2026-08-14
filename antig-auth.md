# Antigravity OAuth, Gemini 3.7 Flash, Sub2API, and ZCode

This document records the working setup for exposing the Antigravity rollout of
Gemini 3.7 Flash through the Sub2API server on VPS2 and using it from ZCode's
custom provider named `2598`.

It intentionally contains no API keys, OAuth access tokens, refresh tokens,
account emails, client secrets, or thought signatures.

## Current working state

| Component | Value |
| --- | --- |
| Sub2API branch | `codex/gemini-3.7-flash` |
| Implementation commit | `938964783a6f9b621b6eac10a6673687c19784cd` |
| VPS2 image | `oghyhk/sub2api:gemini-3.7-9389647` |
| Public model | `gemini-3.7-flash` |
| Upstream Antigravity model | `gemini-3.7-flash-high` |
| Antigravity client identity | `antigravity/2.8.1 windows/amd64` |
| Daily endpoint | `https://daily-cloudcode-pa.googleapis.com` |
| ZCode provider ID | `google-vps2` |
| ZCode provider name | `2598` |
| ZCode base URL | `https://api.k2598.com/v1` |

The public alias deliberately maps to the High variant because the previous
`gemini-3.6-flash` alias also mapped to its High upstream variant.

## What actually made 3.7 work

The failure was caused by stale Antigravity request identity, not account
eligibility.

The old combination was:

- endpoint: `https://daily-cloudcode-pa.sandbox.googleapis.com`
- user-agent version: `2.0.6`
- result: 24 advertised models and no Gemini 3.7 variants

The working combination is:

- endpoint: `https://daily-cloudcode-pa.googleapis.com`
- user-agent version: `2.8.1`
- result: 27 advertised models, including:
  - `gemini-3.7-flash-high`
  - `gemini-3.7-flash-medium`
  - `gemini-3.7-flash-low`

A direct generation request through an existing pool account using
`gemini-3.7-flash-high` returned:

- `modelVersion`: `gemini-3.7-flash`
- `finishReason`: `STOP`
- expected deterministic output

This proved availability before any production or ZCode configuration was
changed.

Antigravity 2.8.1 used internal enum values 1298, 1299, and 1300 for High,
Medium, and Low during discovery. These numbers are diagnostic observations,
not a stable API. Never hardcode them; use the model strings.

## Upstream research

The inspected upstream repository was:

- <https://github.com/NoeFabris/opencode-antigravity-auth>

It was archived on July 17, 2026 and documented generic Gemini 3 Flash/3.1
routing, but did not provide the missing Gemini 3.7 Sub2API configuration.

Google's public Gemini API model catalog still documented Gemini 3.6 Flash at
the time of this setup:

- <https://ai.google.dev/gemini-api/docs/models>

Therefore, Gemini 3.7 in this setup is an Antigravity account rollout. Do not
assume that the same model string is available through AI Studio or the public
Gemini API unless that provider's own model catalog advertises it.

## Local Antigravity verification

The local Antigravity desktop installation was upgraded from 2.3.0 to 2.8.1.
Before installation, the updater package was checked against the updater
manifest and its Windows Authenticode signature was verified as valid and
signed by Google LLC.

After launch, inspect the live model state rather than inferring availability
from documentation. The account must advertise all desired model strings. A
zero-output model-list request is not sufficient proof of generation support;
also run a small real generation with at least 256 output tokens because the
model may spend a small budget entirely on thinking.

## Sub2API source changes

### Antigravity endpoint and user agent

`backend/internal/pkg/antigravity/oauth.go` defines:

```go
DefaultUserAgentVersion = "2.8.1"
antigravityDailyBaseURL = "https://daily-cloudcode-pa.googleapis.com"
```

`backend/internal/pkg/antigravity/client.go` also sets the daily request host to
`daily-cloudcode-pa.googleapis.com` when rewriting the forwarded request.

Both locations matter. Updating only the base URL while leaving a stale Host
header can still route the request incorrectly.

### Canonical model mappings

`backend/internal/domain/constants.go` contains:

```go
"gemini-3.7-flash":        "gemini-3.7-flash-high",
"gemini-3.7-flash-high":   "gemini-3.7-flash-high",
"gemini-3.7-flash-medium": "gemini-3.7-flash-medium",
"gemini-3.7-flash-low":    "gemini-3.7-flash-low",
```

`backend/internal/service/account.go` ensures these canonical defaults are
merged into Antigravity accounts even when an account already has a custom
mapping. This is important: treating an unknown public alias as an identity
mapping would send `gemini-3.7-flash` upstream, but Antigravity advertises the
explicit High/Medium/Low model names.

### Required tests

Run from the `backend` directory:

```powershell
go test -p 1 ./internal/domain ./internal/pkg/antigravity

go test -tags=unit -p 1 ./internal/service `
  -run 'TestAntigravityGatewayService_GetMappedModel'

go test -tags=unit -p 1 ./internal/service `
  -run 'Test(GenerateSessionHash|EnsureGeminiFunctionCallIDs|CachePrefix|EnsureGeminiUpstreamSession|ScopedClient|AntigravityWeeklyWarmup|HandleSmartRetry_ExactlyAtThreshold)'

go test -tags=unit -p 1 ./internal/service -run 'TestDigestSession'
```

The cache-affinity rules in `AGENTS.md` and
`docs/STICKY_SESSION_INVARIANT.md` still apply. Model mapping must not change
session identity, randomly rewrite prior turns, or dislodge a healthy sticky
account.

## VPS2 deployment

### Paths and containers

```text
Source:  /home/sub2api/src/sub2api-gemini-chat-signatures-deploy
Stack:   /home/sub2api/app/rehearsal
Compose: /home/sub2api/app/rehearsal/compose.yaml
App:     sub2api-rehearsal-sub2api-1
DB:      sub2api-rehearsal-db-1
Redis:   sub2api-rehearsal-redis-1
Port:    127.0.0.1:18080 -> container 8080
```

The Docker daemon is rootless under user `sub2api`. Repository ownership and
Docker ownership differ, so run Git as the repository owner and Docker commands
as `sub2api`.

### 1. Build the image

After the branch is pushed and the VPS source checkout is on the exact commit:

```bash
cd /home/sub2api/src/sub2api-gemini-chat-signatures-deploy
sudo -iu sub2api docker build -t oghyhk/sub2api:gemini-3.7-9389647 .
sudo -iu sub2api docker image inspect \
  oghyhk/sub2api:gemini-3.7-9389647 --format '{{.Id}} {{.Created}}'
```

Never replace the live compose image before the new image inspection succeeds.

### 2. Make a private rollback backup

Use a new timestamped directory and a restrictive umask:

```bash
umask 077
backup=/home/sub2api/app/rehearsal/backups/pre-gemini37-YYYYMMDDTHHMMSSZ
install -d -m 700 "$backup"
cp /home/sub2api/app/rehearsal/compose.yaml "$backup/compose.yaml"
sudo -iu sub2api docker exec sub2api-rehearsal-db-1 \
  pg_dump -U postgres -d sub2api -Fc > "$backup/sub2api.dump"
chmod 600 "$backup/compose.yaml" "$backup/sub2api.dump"
stat -c '%n bytes=%s mode=%a' "$backup"/*
```

The backup used for the original rollout is:

```text
/home/sub2api/app/rehearsal/backups/pre-gemini37-20260814T015402Z
```

Its directory is mode 0700 and both files are mode 0600.

### 3. Update runtime settings transactionally

Before changing anything, verify the current values:

```sql
SELECT key, value
FROM settings
WHERE key = 'antigravity_user_agent_version';

SELECT id, models_list_config
FROM groups
WHERE id = 2;

SELECT id, credentials->'model_mapping'
FROM accounts
WHERE platform = 'antigravity'
  AND type = 'oauth'
  AND deleted_at IS NULL;
```

Then update the user-agent version, provider catalog, and account mappings in
one guarded transaction. The production rollout expected exactly eight pool
accounts; adjust that assertion only after independently verifying the pool.

```sql
BEGIN;

DO $deploy$
DECLARE affected integer;
BEGIN

UPDATE settings
SET value = '2.8.1', updated_at = NOW()
WHERE key = 'antigravity_user_agent_version'
  AND value = '2.0.6';

GET DIAGNOSTICS affected = ROW_COUNT;
IF affected <> 1 THEN
  RAISE EXCEPTION 'expected one UA setting row, updated %', affected;
END IF;

UPDATE groups
SET models_list_config = jsonb_set(
      models_list_config,
      '{models}',
      '["gemini-3.7-flash","claude-opus-4-6","claude-opus-4-6-thinking","gemini-3.1-flash-lite","deepseek/deepseek-v4-flash","gpt-5.6-luna"]'::jsonb,
      true
    ),
    updated_at = NOW()
WHERE id = 2
  AND models_list_config->'models' @> '["gemini-3.6-flash"]'::jsonb
  AND NOT (models_list_config->'models' @> '["gemini-3.7-flash"]'::jsonb);

GET DIAGNOSTICS affected = ROW_COUNT;
IF affected <> 1 THEN
  RAISE EXCEPTION 'expected one provider group row, updated %', affected;
END IF;

UPDATE accounts
SET credentials = jsonb_set(
      credentials,
      '{model_mapping}',
      (
        (COALESCE(credentials->'model_mapping', '{}'::jsonb)
          - 'gemini-3.6-flash')
        || '{
          "gemini-3.7-flash":"gemini-3.7-flash-high",
          "gemini-3.7-flash-high":"gemini-3.7-flash-high",
          "gemini-3.7-flash-medium":"gemini-3.7-flash-medium",
          "gemini-3.7-flash-low":"gemini-3.7-flash-low"
        }'::jsonb
      ),
      true
    ),
    updated_at = NOW()
WHERE platform = 'antigravity'
  AND type = 'oauth'
  AND deleted_at IS NULL
  AND credentials->'model_mapping'->>'gemini-3.6-flash'
      = 'gemini-3.6-flash-high';

GET DIAGNOSTICS affected = ROW_COUNT;
IF affected <> 8 THEN
  RAISE EXCEPTION 'expected eight Antigravity mappings, updated %', affected;
END IF;

END
$deploy$;

COMMIT;
```

The final guard value must match the independently verified pool size. It was
eight for this rollout; do not blindly reuse that number if the pool changes.

Post-migration checks must show:

- setting value `2.8.1`
- group 2 contains `gemini-3.7-flash`
- group 2 does not contain `gemini-3.6-flash`
- every Antigravity OAuth account maps the public 3.7 alias to High
- no account retains the old 3.6 alias

### 4. Recreate only the app service

Change the `sub2api` image in `compose.yaml` only after verifying the old line
matches exactly:

```yaml
services:
  sub2api:
    image: oghyhk/sub2api:gemini-3.7-9389647
    pull_policy: never
```

Then recreate only the app container:

```bash
sudo -iu sub2api bash -lc \
  'cd /home/sub2api/app/rehearsal && docker compose up -d --no-deps sub2api'
```

Do not restart PostgreSQL or Redis for this change.

### 5. Verify production health

```bash
sudo -iu sub2api docker inspect sub2api-rehearsal-sub2api-1 \
  --format 'image={{.Config.Image}} health={{.State.Health.Status}} restart={{.RestartCount}}'

curl -fsS http://127.0.0.1:18080/health

sudo -iu sub2api bash -lc \
  'cd /home/sub2api/app/rehearsal && docker compose ps'
```

Expected state:

```text
image=oghyhk/sub2api:gemini-3.7-9389647
health=healthy
restart=0
{"status":"ok"}
```

## ZCode provider 2598

ZCode stores its custom provider configuration here on this workstation:

```text
C:\Users\oghyh\.zcode\v2\config.json
```

The application may keep configuration in memory. Close ZCode completely,
including its tray/background processes, before editing, or verify afterward
that shutdown did not overwrite the file. Always make a hash-checked backup.

The working provider metadata is:

```json
{
  "name": "2598",
  "kind": "openai-compatible",
  "options": {
    "baseURL": "https://api.k2598.com/v1",
    "apiKey": "<existing secret; never commit or paste into this document>"
  },
  "models": {
    "gemini-3.7-flash": {
      "name": "Gemini 3.7 Flash",
      "reasoning": {
        "enabled": true,
        "variants": ["high"],
        "defaultVariant": "high"
      },
      "limit": {
        "context": 1048576,
        "output": 65536
      },
      "modalities": {
        "input": ["text", "image", "pdf", "audio"],
        "output": ["text"]
      },
      "zcode": {
        "modified": true,
        "priority": 100
      }
    }
  }
}
```

Replace the `gemini-3.6-flash` model key and display name; do not leave both
models behind when the goal is a replacement. Preserve all unrelated provider
models and the existing API key.

After restart, the ZCode runtime log should report both:

```text
modelCurrent: google-vps2/gemini-3.7-flash
provider/model ready: modelId=gemini-3.7-flash providerId=google-vps2
```

The log directory is:

```text
C:\Users\oghyh\.zcode\v2\logs
```

## End-to-end tests

### Public model catalog

Use the existing ZCode provider key without printing it:

```powershell
$headers = @{ Authorization = "Bearer $env:SUB2API_TEST_KEY" }
$catalog = Invoke-RestMethod `
  -Uri 'https://api.k2598.com/v1/models' `
  -Headers $headers
$ids = @($catalog.data.id)

if ($ids -notcontains 'gemini-3.7-flash') {
  throw '3.7 is missing from the public catalog'
}
if ($ids -contains 'gemini-3.6-flash') {
  throw 'stale 3.6 is still exposed'
}
```

### Public chat completion

```powershell
$body = @{
  model = 'gemini-3.7-flash'
  messages = @(
    @{ role = 'user'; content = 'Return exactly PUBLIC37OK and nothing else.' }
  )
  max_tokens = 512
  stream = $false
} | ConvertTo-Json -Depth 8 -Compress

$response = Invoke-RestMethod `
  -Method Post `
  -Uri 'https://api.k2598.com/v1/chat/completions' `
  -Headers $headers `
  -ContentType 'application/json' `
  -Body $body `
  -TimeoutSec 180

if ($response.choices[0].message.content.Trim() -ne 'PUBLIC37OK') {
  throw 'unexpected model response'
}
```

The production verification returned:

```text
requested model: gemini-3.7-flash
response model:  gemini-3.7-flash
finish reason:   stop
content:         PUBLIC37OK
```

### ZCode in-app test

1. Open a new ZCode task.
2. Confirm the model selector displays `2598/gemini-3.7-flash`.
3. Send: `Return exactly ZCODE37OK and nothing else.`
4. Confirm the assistant returns only `ZCODE37OK`.

The original in-app test completed successfully in 17 seconds.

### Server-side corroboration

Inspect only the recent, sanitized model tokens from the new container logs.
The same test window must contain both:

```text
gemini-3.7-flash
gemini-3.7-flash-high
```

This confirms the public alias was received and the request was routed to the
High upstream variant. Do not dump complete debug logs into tickets or chat;
they can contain sensitive request metadata.

## Rollback

Prefer a targeted rollback over restoring the full database dump:

1. Replace the compose image with the previously recorded image.
2. Change `antigravity_user_agent_version` back to its recorded prior value.
3. Replace `gemini-3.7-flash` with `gemini-3.6-flash` in group 2.
4. Remove the four 3.7 mapping keys from each Antigravity account.
5. Restore `gemini-3.6-flash -> gemini-3.6-flash-high`.
6. Recreate only the `sub2api` service.
7. Re-run local and public health checks.

Use the PostgreSQL dump only if a targeted inverse transaction cannot restore
the prior state. A full restore has a larger blast radius and requires a
separate maintenance plan.

For ZCode, restore the timestamped `config.json.pre-gemini37-*` backup and
restart ZCode completely.

## Security and operational notes

- Never commit or print ZCode's `apiKey`.
- Never log Antigravity access tokens, refresh tokens, client secrets, account
  emails, or thought signatures.
- Keep database dumps in a mode-0700 directory with mode-0600 files.
- Keep the Antigravity account pool and session affinity behavior unchanged
  unless a separately reviewed routing change requires it.
- A model-list response proves advertisement, not successful generation. Always
  run a real, bounded request before rollout.
- An individual account failure does not prove the model is unavailable. Test
  the eligible pool and distinguish auth, quota, overload, and model errors.
- The upstream authentication project warns that unofficial Antigravity proxy
  usage may violate Google's terms and may put accounts at risk. Use only
  accounts whose owner has accepted that risk.

## Future model-upgrade checklist

1. Upgrade or inspect the current signed Antigravity client.
2. Enumerate models with the current endpoint and user-agent identity.
3. Prove generation using one eligible account before editing production.
4. Add canonical base and variant mappings with tests.
5. Run cache-affinity and signature/session tests.
6. Commit and push the exact source revision.
7. Build a uniquely tagged image and inspect it.
8. Create and permission-check compose and database backups.
9. Update catalog and account mappings transactionally with row-count guards.
10. Recreate only the application service.
11. Verify local health, public catalog, public completion, ZCode runtime state,
    an in-app completion, upstream mapping in sanitized logs, and restart count.
12. Keep the rollback image and database backup until the new route has remained
    stable through normal production traffic.
