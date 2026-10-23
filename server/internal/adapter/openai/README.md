# internal/adapter/openai

**Clean Architecture layer: Interface Adapter (outbound/driven).** Implements the `Extractor` port.

**Role.** Implements `transaction_attachment_extraction.Extractor` against the **OpenAI Chat Completions wire format** (`POST {base_url}/chat/completions`, vision content parts, `response_format` json_schema structured output) — the **only** place that speaks HTTP to an external vision provider. Owns `openai.Config` (`OPENAI_*`).

## Files

- `openai.go` — package documentation, `Client` + `NewClient(cfg)`, and the shared vision request mechanics (`visionParams`, `structuredContent`, provider-error diagnostics). It contains no flow-specific prompt or schema.
- `health.go` — health `Check` via the configured model endpoint.
- `transaction_prefill.go` — the receipt extraction flow: `ExtractTransactionPrefill`, its strict schema, response DTO, and decoder. Adding another flow means adding another method/input contract here, never a prompt switch.
- `transaction_prefill_prompt.go` — the receipt flow's complete instruction. It states every expected field and passes the user's active **expense categories** and **income categories** as separate explicit lists.
- `config.go` — `Config{ APIKey, BaseURL, Model, TimeoutSeconds }` (`OPENAI_API_KEY` required; `OPENAI_BASE_URL` default `https://api.openai.com/v1`; `OPENAI_EXTRACTION_MODEL` default `gpt-4.1-mini`; `OPENAI_TIMEOUT_SECONDS` default `300`); `LoadConfig()`.

## Behavior

- Each flow sends **exactly one** user message: its instruction text part plus one `image_url` content part per supplied URL, **verbatim and in order**. URLs are our short-lived attachment-content grants (`{PUBLIC_API_BASE_URL}/api/v1/attachment-content?grant=…`) — the provider fetches the bytes itself; **no attachment bytes ever pass through this adapter**.
- The request body is fixed by `visionParams`: configured model, `temperature: 0`, and `response_format: {type:"json_schema", json_schema:{name, schema, strict:true}}`.
- The receipt prompt asks for type, total amount, currency, date, optional visible time (`HH:MM`), name, description, type-specific category hint, and line items with non-empty labels. Line-item amounts may be negative for discounts/coupons; zero is not allowed. Every present scalar and line item must include an integer confidence from 0 to 100. Category names are JSON-encoded and explicitly separated into `expense_categories` and `income_categories`.
- **Every failure returns a plain error, never a domain error:** transport error, timeout, non-2xx (including `429`/`5xx`), `content_filter`, empty content, or invalid structured JSON. The service maps it to the 503 kind. Provider 400 diagnostics are retained in the wrapped error for internal boundary logs only.
- The SDK client carries the API key, base URL, an HTTP client bounded by `TimeoutSeconds`, and **no retries**.

## Rules

- **Protocol, not a vendor SDK surface.** Switching providers is a config-only change (`OPENAI_BASE_URL` + `OPENAI_API_KEY` + model); the `Extractor` port keeps that swap safe.
- **One flow, one method and prompt file.** Do not add a generic prompt string or a flow switch to the service contract.
- **Images-only is enforced by the service.** The adapter never inspects or trusts the media type.
- **Adapters don't log.** They return errors and let the service/caller log outcomes; never log attachment bytes, content URLs, filenames, or raw responses.
- Imports: stdlib, `domain`, `service/transaction_attachment_extraction`, `utils/configuration`, and the OpenAI-compatible SDK. Nothing else internal.

## Testing requirement for this directory

Black-box (`package openai_test`) against `httptest.NewServer` — never a mocked `http.Client`. Cover request shape, explicit per-type category lists, ordered multi-image parts, `occurred_at`, error mapping, base-URL portability, bounded timeout, and provider diagnostics.
