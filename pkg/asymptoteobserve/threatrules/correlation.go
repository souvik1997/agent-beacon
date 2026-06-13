package threatrules

import (
	"time"

	"github.com/asymptote-labs/agent-beacon/pkg/asymptoteobserve"
)

// evaluateCorrelation runs the ordered-window state machine. Events are grouped by
// session.id (scope: session) and each group is evaluated independently; the rule
// matches if any session satisfies the sequence.
func (c *CompiledRule) evaluateCorrelation(events []asymptoteobserve.Event) (Verdict, error) {
	groups, order := groupBySession(events)
	for _, sid := range order {
		matched, err := c.matchSession(groups[sid])
		if err != nil {
			return "", err
		}
		if matched {
			return VerdictMatch, nil
		}
	}
	return VerdictNoMatch, nil
}

// matchSession reports whether one session's ordered events satisfy the step sequence
// within the window. Steps must occur in order; the elapsed time from the first matched
// step to the final matched step must not exceed the window. Timestamps are RFC3339; if a
// step's timestamp is absent the window is not enforced against it (positive fixtures need
// not carry timestamps, while window fixtures supply them).
func (c *CompiledRule) matchSession(events []asymptoteobserve.Event) (bool, error) {
	stepIdx := 0
	var startTime time.Time
	startKnown := false
	final := len(c.steps) - 1

	for i := range events {
		matched, err := EvalMatch(c.steps[stepIdx], events[i])
		if err != nil {
			return false, err
		}
		if !matched {
			continue
		}
		t, known := eventTime(events[i])
		if stepIdx == 0 {
			startTime, startKnown = t, known
		}
		if stepIdx == final {
			if startKnown && known && t.Sub(startTime) > c.window {
				// This alignment completes outside the window; abandon it and keep
				// scanning later events for a fresh sequence.
				stepIdx = 0
				startKnown = false
				continue
			}
			return true, nil
		}
		stepIdx++
	}
	return false, nil
}

// groupBySession partitions events by session.id, preserving per-group input order and
// recording the order in which sessions first appear for deterministic evaluation.
func groupBySession(events []asymptoteobserve.Event) (map[string][]asymptoteobserve.Event, []string) {
	groups := make(map[string][]asymptoteobserve.Event)
	var order []string
	for i := range events {
		sid := ""
		if events[i].Session != nil {
			sid = events[i].Session.ID
		}
		if _, seen := groups[sid]; !seen {
			order = append(order, sid)
		}
		groups[sid] = append(groups[sid], events[i])
	}
	return groups, order
}

// eventTime parses an event's RFC3339 timestamp. The bool is false when the timestamp is
// absent or unparseable, in which case window enforcement is skipped for that event.
func eventTime(e asymptoteobserve.Event) (time.Time, bool) {
	if e.Timestamp == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, e.Timestamp)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
