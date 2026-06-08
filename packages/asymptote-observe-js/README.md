# Asymptote Observe TypeScript SDK

Asymptote Observe instruments cloud-hosted agent apps and exports OpenTelemetry
spans that remain compatible with Beacon's normalized JSONL event schema.

## Hosted Asymptote Observe

```bash
npm install @asymptote-labs/observe
export ASYMPTOTE_OBSERVE_API_KEY=...
```

```typescript
import { AsymptoteObserve } from "@asymptote-labs/observe";

AsymptoteObserve.initialize({
  apiKey: process.env.ASYMPTOTE_OBSERVE_API_KEY,
});
```

Hosted Observe contract:

- Transport: OpenTelemetry over HTTP.
- Default base URL: `https://api.asymptotelabs.ai`.
- Observe path: `/v1/observe` is appended when the endpoint does not already include it.
- Auth: `authorization: Bearer <ASYMPTOTE_OBSERVE_API_KEY>`.
- Override base URL with `ASYMPTOTE_OBSERVE_BASE_URL` or `initialize({ baseUrl })`.

The TypeScript SDK preserves this hosted contract, but this repository does not
implement the managed ingest service. Use `baseUrl` with a local stub in tests.

## Local Or Customer-Managed Observe Export

```typescript
import { AsymptoteObserve } from "@asymptote-labs/observe";

AsymptoteObserve.initialize({
  otlpEndpoint: process.env.ASYMPTOTE_OBSERVE_ENDPOINT,
});
```

The SDK appends `/v1/observe` when the endpoint does not already include it.
Cloud spans carry `beacon.origin=cloud` and the compatibility
attributes documented in `docs/asymptote-observe-sdk.md`.

Endpoint precedence:

1. `initialize({ otlpEndpoint })`
2. `ASYMPTOTE_OBSERVE_ENDPOINT`
3. `OTEL_EXPORTER_OTLP_ENDPOINT`
4. hosted `baseUrl` when an API key is present

Beacon endpoint installs remain local-first and never require hosted credentials.

## Automatic AI Instrumentation

Asymptote Observe follows an OpenTelemetry-first pattern: initialize once
as early as possible, then use existing OpenLLMetry instrumentations for common
provider SDKs.

```typescript
import { AsymptoteObserve } from "@asymptote-labs/observe";

AsymptoteObserve.initialize({
  apiKey: process.env.ASYMPTOTE_OBSERVE_API_KEY,
});
```

By default the SDK enables OpenAI and Anthropic OpenLLMetry instrumentations.
Disable or configure them if needed:

```typescript
AsymptoteObserve.initialize({
  apiKey: process.env.ASYMPTOTE_OBSERVE_API_KEY,
  instrumentationOptions: {
    openAI: { traceContent: true },
    anthropic: false,
  },
});
```

If your app already owns OpenTelemetry setup, use
`AsymptoteObserve.instrumentations()` with that provider instead of calling
`initialize()` twice.

## Vercel AI SDK

```typescript
import { registerTelemetry } from "ai";
import { OpenTelemetry } from "@ai-sdk/otel";
import { AsymptoteObserve } from "@asymptote-labs/observe";

AsymptoteObserve.initialize();

registerTelemetry(new OpenTelemetry({
  tracer: AsymptoteObserve.getTracer(),
}));
```

## Claude Agent SDK

```typescript
import { query as originalQuery } from "@anthropic-ai/claude-agent-sdk";
import { AsymptoteObserve } from "@asymptote-labs/observe";

AsymptoteObserve.initialize();

const query = AsymptoteObserve.wrapClaudeAgentQuery(originalQuery);
```

For the first SDK release, prefer provider-level Anthropic/OpenLLMetry
instrumentation or `observe()` around the agent entry point. Custom Claude Agent
SDK hooks are deferred until those paths leave important telemetry gaps.

## Custom Agent Steps

```typescript
const plan = AsymptoteObserve.observe(
  {
    name: "agent.plan",
    attributes: {
      "beacon.harness.name": "custom_agent",
      "beacon.event.action": "tool.invoked",
    },
  },
  async (input: string) => {
    return runPlanner(input);
  },
);
```

Call `AsymptoteObserve.flush()` before short-lived scripts exit, or
`AsymptoteObserve.shutdown()` when the process owns the provider lifecycle.

## Beta Support Matrix

- OpenAI SDK: OpenLLMetry instrumentation enabled by default.
- Anthropic SDK: OpenLLMetry instrumentation enabled by default.
- Vercel AI SDK: register `@ai-sdk/otel` with `AsymptoteObserve.getTracer()`.
- Claude Agent SDK: wrap query functions with `wrapClaudeAgentQuery()` or use `observe()`.
- Custom orchestration: wrap functions with `observe()`.
- Custom Claude hook and OpenAI Agents tracing adapters are intentionally deferred.
