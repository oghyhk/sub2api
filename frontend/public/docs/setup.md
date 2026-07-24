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

## AI-agent instruction

Paste this page URL into your AI agent with:

```text
Read this Sub2API setup guide and configure my current tool. Ask me for the API key only when you are ready to store it locally. Use the correct GPT/OpenAI-compatible or native Gemini protocol, preserve my existing configuration, and verify with a small request.
```
