package beaconevent

import (
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestEventsFromTracesNormalizesObserveSDKSpan(t *testing.T) {
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	resourceAttrs := resourceSpans.Resource().Attributes()
	resourceAttrs.PutStr("beacon.origin", "cloud")
	resourceAttrs.PutStr("beacon.harness.name", "asymptote_observe")
	resourceAttrs.PutStr("service.name", "agent-api")

	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	scopeSpans.Scope().SetName("asymptote-observe")
	span := scopeSpans.Spans().AppendEmpty()
	span.SetName("agent.plan")
	span.SetKind(ptrace.SpanKindClient)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Unix(1700000000, 0).UTC()))
	attrs := span.Attributes()
	attrs.PutStr("beacon.event.action", "prompt.submitted")
	attrs.PutStr("beacon.event.category", "prompt")
	attrs.PutStr("beacon.prompt.text", "summarize this deployment")
	attrs.PutStr("gen_ai.provider.name", "openai")
	attrs.PutStr("gen_ai.operation.name", "chat")
	attrs.PutStr("gen_ai.request.model", "gpt-4o-mini")
	attrs.PutInt("gen_ai.usage.input_tokens", 12)
	attrs.PutInt("gen_ai.usage.output_tokens", 34)

	events := NewConverter(Options{}).EventsFromTraces(traces)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	event := events[0]
	if event.Origin != "cloud" {
		t.Fatalf("origin = %q, want cloud", event.Origin)
	}
	if event.Harness.Name != "asymptote_observe" {
		t.Fatalf("harness = %q, want asymptote_observe", event.Harness.Name)
	}
	if event.Event.Action != "prompt.submitted" {
		t.Fatalf("action = %q, want prompt.submitted", event.Event.Action)
	}
	if event.Event.Category != "prompt" {
		t.Fatalf("category = %q, want prompt", event.Event.Category)
	}
	if event.Prompt == nil || event.Prompt.Text != "summarize this deployment" {
		t.Fatalf("prompt = %#v, want captured prompt text", event.Prompt)
	}
	if event.GenAI == nil || event.GenAI.Provider == nil || event.GenAI.Provider.Name != "openai" {
		t.Fatalf("gen_ai provider = %#v, want openai", event.GenAI)
	}
	if event.GenAI.Request == nil || event.GenAI.Request.Model != "gpt-4o-mini" {
		t.Fatalf("gen_ai request = %#v, want model", event.GenAI.Request)
	}
	if event.GenAI.Usage == nil || event.GenAI.Usage.InputTokens == nil || *event.GenAI.Usage.InputTokens != 12 {
		t.Fatalf("gen_ai usage input = %#v, want 12", event.GenAI.Usage)
	}
	if event.GenAI.Usage.OutputTokens == nil || *event.GenAI.Usage.OutputTokens != 34 {
		t.Fatalf("gen_ai usage output = %#v, want 34", event.GenAI.Usage)
	}
}
