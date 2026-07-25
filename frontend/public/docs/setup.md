# Sub2API General API Setup

Use this guide to configure any tool that supports an OpenAI-compatible API or the native Google Gemini API.

> AI-agent setup: paste this page URL into your AI agent and ask it to configure your tool. Never paste your API key into a public chat or issue.

## Connection details

| Purpose | Value |
| --- | --- |
| Service root | `https://api.k2598.com` |
| OpenAI-compatible Base URL | `https://api.k2598.com/v1` |
| Native Gemini Base URL | `https://api.k2598.com/antigravity/v1beta` |
| OpenAI-compatible authentication | `Authorization: Bearer YOUR_API_KEY` |
| Native Gemini authentication | `x-goog-api-key: YOUR_API_KEY` |

Create a key in **API Keys**, then click **Set Up** beside that key for tool-specific instructions.

## Which protocol should I use?

- **GPT models:** use the OpenAI-compatible Base URL and OpenAI request format.
- **Gemini models in OpenAI-compatible tools:** use the OpenAI-compatible Base URL and OpenAI request format.
- **Native Gemini tools/SDKs:** use the Native Gemini Base URL, the `x-goog-api-key` header, and Gemini's `generateContent` request format.
- Do not send native Gemini requests to `/v1/chat/completions`, and do not send OpenAI chat payloads to `:generateContent`.

## Supported endpoints

### OpenAI-compatible

- `GET https://api.k2598.com/v1/models`
- `POST https://api.k2598.com/v1/chat/completions`
- `POST https://api.k2598.com/v1/responses`

### Native Gemini

- `POST https://api.k2598.com/antigravity/v1beta/models/{model}:generateContent`

## Supported model IDs

| Model ID | Family | Preferred protocol | Context window |
| --- | --- | --- | ---: |
| `gpt-5.6-sol` | GPT | OpenAI-compatible API | 358,000 |
| `gpt-5.6-terra` | GPT | OpenAI-compatible API | 358,000 |
| `gpt-5.6-luna` | GPT | OpenAI-compatible API | 358,000 |
| `gemini-3.1-pro` | Gemini | Google Gemini API | 1,048,576 |
| `gemini-3.6-flash` | Gemini | Google Gemini API | 1,048,576 |

## Minimal OpenAI-compatible example

```bash
curl "https://api.k2598.com/v1/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.6-luna","messages":[{"role":"user","content":"Reply with OK"}]}'
```

The same OpenAI-compatible endpoint can be used with a Gemini model when your tool only supports the OpenAI protocol:

```bash
curl "https://api.k2598.com/v1/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"Reply with OK"}]}'
```

## Minimal native Gemini example

```bash
curl "https://api.k2598.com/antigravity/v1beta/models/gemini-3.6-flash:generateContent" \
  -H "x-goog-api-key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"parts":[{"text":"Reply with OK"}]}]}'
```

## Generic tool setup checklist

1. Create an API key in the Sub2API dashboard.
2. Open the tool's custom provider, OpenAI-compatible provider, or Gemini provider settings.
3. Enter the matching Base URL from this guide.
4. Enter the API key in the matching authentication field.
5. Enter one of the model IDs exactly as shown above.
6. Save, start a new session, and send a small test request.
7. If the tool appends `/v1` automatically, give it `https://api.k2598.com`; otherwise use `https://api.k2598.com/v1`.


### Setting up ChatGPT desktop app (Codex Desktop)

If you are using the ChatGPT desktop app (Codex Desktop), you can pass one of these exact prompts to an AI agent to set it up for you:

**For Windows:**
```text
Please help me set up the ChatGPT desktop app (Codex Desktop) on Windows to use my Sub2API key.

My API base URL is: https://api.k2598.com/v1
My Native Gemini Base URL is: https://api.k2598.com/antigravity/v1beta
My API Key is: [Replace with your API Key]

Please follow the Sub2API Codex Desktop architecture:
1. Prevent OpenAI model discovery: keep Codex signed out, set `model_provider = "sub2api"` in `~/.codex/config.toml`, use an absolute local `model_catalog_json` path, and set `[model_providers.sub2api] requires_openai_auth = false`.
2. LiteLLM bridge (`127.0.0.1:4011`): Create 5 upstream routes in `~/.codex/providers/sub2api-bridge/config.yaml`: `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`, `gpt-5.4` (mapped to `gemini/gemini-3.1-pro`), and `gemini-3.6-flash`.
3. Model catalog (`~/.codex/providers/sub2api-bridge/model-catalog.json`): Add all 5 entries. For `gpt-5.4`, set display name to "Gemini 3.1 Pro" and effort to "high".
4. Set up an on-demand bridge launcher (`ensure-bridge.ps1`) and runner (`bridge-runner.py`). Configure `[model_providers.sub2api.auth]` in `config.toml` to call `powershell.exe -NoProfile -ExecutionPolicy Bypass -File ... ensure-bridge.ps1`.
5. Credentials: Store the Sub2API key and a random bridge token as Windows user environment variables (`SUB2API_API_KEY`, `CODEX_SUB2API_BRIDGE_KEY`). Do not write them into any config files! The bridge launcher must pass them to the bridge and return the bridge token to Codex.
6. Subagent: In `config.toml`, set `[agents] default_subagent_model = "gemini-3.6-flash"` and medium effort. Set up `~/.codex/agents/gemini_flash.toml` to select it directly.
7. Restart: Stop the port 4011 process, stop every `OpenAI.Codex` MSIX process (including `ChatGPT.exe` host and backend), then relaunch.
8. Validate using active `CODEX_CLI_PATH`: `debug models` must return exactly the 5 local entries.
```

**For macOS:**
```text
Please help me set up the ChatGPT desktop app (Codex Desktop) on macOS to use my Sub2API key.

My API base URL is: https://api.k2598.com/v1
My Native Gemini Base URL is: https://api.k2598.com/antigravity/v1beta
My API Key is: [Replace with your API Key]

Please follow the Sub2API Codex Desktop architecture:
1. Prevent OpenAI model discovery: keep Codex signed out, set `model_provider = "sub2api"` in `~/.codex/config.toml`, use an absolute local `model_catalog_json` path, and set `[model_providers.sub2api] requires_openai_auth = false`.
2. LiteLLM bridge (`127.0.0.1:4011`): Create 5 upstream routes in `~/.codex/providers/sub2api-bridge/config.yaml`: `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`, `gpt-5.4` (mapped to `gemini/gemini-3.1-pro`), and `gemini-3.6-flash`.
3. Model catalog (`~/.codex/providers/sub2api-bridge/model-catalog.json`): Add all 5 entries. For `gpt-5.4`, set display name to "Gemini 3.1 Pro" and effort to "high".
4. Set up an on-demand bridge launcher (`ensure-bridge.sh`) and runner (`bridge-runner.py`). Configure `[model_providers.sub2api.auth]` in `config.toml` to call `bash ... ensure-bridge.sh`.
5. Credentials: Store the Sub2API key and a random bridge token securely in the macOS user environment (e.g. `~/.zshenv` or LaunchAgents) so the bridge launcher can read them. Do not write them into any config files! The bridge launcher must pass them to the bridge and return the bridge token to Codex.
6. Subagent: In `config.toml`, set `[agents] default_subagent_model = "gemini-3.6-flash"` and medium effort. Set up `~/.codex/agents/gemini_flash.toml` to select it directly.
7. Restart: Stop the port 4011 process, fully quit the Codex Desktop app (Cmd+Q) and all its background processes, then relaunch.
8. Validate using active `CODEX_CLI_PATH`: `debug models` must return exactly the 5 local entries.
```

## AI-agent instruction

Paste this page URL into your AI agent with:

```text
Read this Sub2API setup guide and configure my current tool. Ask me for the API key only when you are ready to store it locally. Use the correct GPT/OpenAI-compatible or native Gemini protocol, preserve my existing configuration, and verify with a small request.
```
