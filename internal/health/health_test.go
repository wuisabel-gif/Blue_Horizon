package health

import (
	"strings"
	"testing"

	"blue-horizon/internal/analyzer"
)

func TestAssessHealthyWhenNoEvents(t *testing.T) {
	assessment := Assess(analyzer.Result{})

	if assessment.Verdict != VerdictHealthy {
		t.Fatalf("expected healthy verdict, got %s", assessment.Verdict)
	}
	if len(assessment.Counts) != 0 {
		t.Fatalf("expected no counts, got %#v", assessment.Counts)
	}
	if len(assessment.Hints) != 0 {
		t.Fatalf("expected no hints, got %#v", assessment.Hints)
	}
}

func TestAssessDangerAndFrameHintForNearHalfTurnMismatch(t *testing.T) {
	result := analyzer.Result{
		MaxMismatchDeg: 180.11,
		Events: []analyzer.Event{
			{Severity: analyzer.Danger, Rule: "estimator_roll_mismatch", ValueDeg: 180.11},
		},
	}

	assessment := Assess(result)

	if assessment.Verdict != VerdictDanger {
		t.Fatalf("expected danger verdict, got %s", assessment.Verdict)
	}
	if len(assessment.Counts) != 1 {
		t.Fatalf("expected one count, got %#v", assessment.Counts)
	}
	if got := FormatRuleCount(assessment.Counts[0]); got != "DANGER estimator_roll_mismatch: 1 events" {
		t.Fatalf("unexpected count line %q", got)
	}
	if len(assessment.Hints) != 1 {
		t.Fatalf("expected one frame hint, got %#v", assessment.Hints)
	}
	if !strings.Contains(assessment.Hints[0], "IMU-to-base_link") {
		t.Fatalf("expected frame-transform hint, got %q", assessment.Hints[0])
	}
}
