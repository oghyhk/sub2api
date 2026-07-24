import { PRODUCT_MODELS } from './models'

export interface GeneralSetupUrls {
  serviceRoot: string
  openAIBase: string
  nativeGeminiBase: string
}

export function resolveGeneralSetupUrls(baseUrl: string): GeneralSetupUrls {
  const serviceRoot = baseUrl
    .trim()
    .replace(/\/antigravity\/v1beta\/?$/, '')
    .replace(/\/v1\/?$/, '')
    .replace(/\/+$/, '')

  return {
    serviceRoot,
    openAIBase: `${serviceRoot}/v1`,
    nativeGeminiBase: `${serviceRoot}/antigravity/v1beta`,
  }
}

export function getProductModelIds(providerFamily?: 'gpt' | 'gemini'): string[] {
  return PRODUCT_MODELS
    .filter((model) => !providerFamily || model.providerFamily === providerFamily)
    .map((model) => model.id)
}

export function buildOpenAICurlExample(openAIBase: string, apiKey = 'YOUR_API_KEY'): string {
  return `curl "${openAIBase}/chat/completions" \\
  -H "Authorization: Bearer ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-5.6-luna","messages":[{"role":"user","content":"Reply with OK"}]}'`
}

export function buildNativeGeminiCurlExample(
  nativeGeminiBase: string,
  apiKey = 'YOUR_API_KEY',
): string {
  return `curl "${nativeGeminiBase}/models/gemini-3.6-flash:generateContent" \\
  -H "x-goog-api-key: ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{"contents":[{"parts":[{"text":"Reply with OK"}]}]}'`
}

export function buildOpenAIPowerShellExample(
  openAIBase: string,
  apiKey = 'YOUR_API_KEY',
): string {
  return `$headers = @{ Authorization = "Bearer ${apiKey}" }
$body = @{ model = "gpt-5.6-luna"; messages = @(@{ role = "user"; content = "Reply with OK" }) } | ConvertTo-Json -Depth 5
Invoke-RestMethod -Method Post -Uri "${openAIBase}/chat/completions" -Headers $headers -ContentType "application/json" -Body $body`
}

export function buildNativeGeminiPowerShellExample(
  nativeGeminiBase: string,
  apiKey = 'YOUR_API_KEY',
): string {
  return `$headers = @{ "x-goog-api-key" = "${apiKey}" }
$body = @{ contents = @(@{ parts = @(@{ text = "Reply with OK" }) }) } | ConvertTo-Json -Depth 6
Invoke-RestMethod -Method Post -Uri "${nativeGeminiBase}/models/gemini-3.6-flash:generateContent" -Headers $headers -ContentType "application/json" -Body $body`
}

export function buildGeneralConnectionDetails(baseUrl: string): string {
  const urls = resolveGeneralSetupUrls(baseUrl)
  return [
    `Service root: ${urls.serviceRoot}`,
    `Base URL (OpenAI-compatible): ${urls.openAIBase}`,
    `Base URL (native Gemini): ${urls.nativeGeminiBase}`,
    'Authentication (OpenAI-compatible): Authorization: Bearer YOUR_API_KEY',
    'Authentication (native Gemini): x-goog-api-key: YOUR_API_KEY',
    'OpenAI-compatible endpoints: POST /chat/completions, POST /responses, GET /models',
    'Native Gemini endpoint: POST /models/{model}:generateContent',
    `GPT models: ${getProductModelIds('gpt').join(', ')}`,
    `Gemini models: ${getProductModelIds('gemini').join(', ')}`,
  ].join('\n')
}

export function buildGeneralSetupMarkdown(baseUrl: string): string {
  const urls = resolveGeneralSetupUrls(baseUrl)
  const modelRows = PRODUCT_MODELS.map((model) => {
    const provider = model.providerFamily === 'gpt' ? 'GPT' : 'Gemini'
    return `| \`${model.id}\` | ${provider} | ${model.protocol} | ${model.contextWindow.toLocaleString('en-US')} |`
  }).join('\n')

  return `# Sub2API General API Setup

Use this guide to configure any tool that supports an OpenAI-compatible API or the native Google Gemini API.

> AI-agent setup: paste this page URL into your AI agent and ask it to configure your tool. Never paste your API key into a public chat or issue.

## Connection details

| Purpose | Value |
| --- | --- |
| Service root | \`${urls.serviceRoot}\` |
| OpenAI-compatible Base URL | \`${urls.openAIBase}\` |
| Native Gemini Base URL | \`${urls.nativeGeminiBase}\` |
| OpenAI-compatible authentication | \`Authorization: Bearer YOUR_API_KEY\` |
| Native Gemini authentication | \`x-goog-api-key: YOUR_API_KEY\` |

Create a key in **API Keys**, then click **Set Up** beside that key for tool-specific instructions.

## Which protocol should I use?

- **GPT models:** use the OpenAI-compatible Base URL and OpenAI request format.
- **Gemini models in OpenAI-compatible tools:** use the OpenAI-compatible Base URL and OpenAI request format.
- **Native Gemini tools/SDKs:** use the Native Gemini Base URL, the \`x-goog-api-key\` header, and Gemini's \`generateContent\` request format.
- Do not send native Gemini requests to \`/v1/chat/completions\`, and do not send OpenAI chat payloads to \`:generateContent\`.

## Supported endpoints

### OpenAI-compatible

- \`GET ${urls.openAIBase}/models\`
- \`POST ${urls.openAIBase}/chat/completions\`
- \`POST ${urls.openAIBase}/responses\`

### Native Gemini

- \`POST ${urls.nativeGeminiBase}/models/{model}:generateContent\`

## Supported model IDs

| Model ID | Family | Preferred protocol | Context window |
| --- | --- | --- | ---: |
${modelRows}

## Minimal OpenAI-compatible example

\`\`\`bash
${buildOpenAICurlExample(urls.openAIBase)}
\`\`\`

The same OpenAI-compatible endpoint can be used with a Gemini model when your tool only supports the OpenAI protocol:

\`\`\`bash
curl "${urls.openAIBase}/chat/completions" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"Reply with OK"}]}'
\`\`\`

## Minimal native Gemini example

\`\`\`bash
${buildNativeGeminiCurlExample(urls.nativeGeminiBase)}
\`\`\`

## Generic tool setup checklist

1. Create an API key in the Sub2API dashboard.
2. Open the tool's custom provider, OpenAI-compatible provider, or Gemini provider settings.
3. Enter the matching Base URL from this guide.
4. Enter the API key in the matching authentication field.
5. Enter one of the model IDs exactly as shown above.
6. Save, start a new session, and send a small test request.
7. If the tool appends \`/v1\` automatically, give it \`${urls.serviceRoot}\`; otherwise use \`${urls.openAIBase}\`.

## AI-agent instruction

Paste this page URL into your AI agent with:

\`\`\`text
Read this Sub2API setup guide and configure my current tool. Ask me for the API key only when you are ready to store it locally. Use the correct GPT/OpenAI-compatible or native Gemini protocol, preserve my existing configuration, and verify with a small request.
\`\`\`
`
}
