package health

import (
	"fmt"
	"sort"
	"strings"

	"blue-horizon/internal/analyzer"
)

type Verdict string

const (
	VerdictHealthy Verdict = "HEALTHY"
	VerdictWarning Verdict = "WARNING"
	VerdictDanger  Verdict = "DANGER"
)

type RuleCount struct {
	Severity analyzer.Severity
	Rule     string
	Count    int
}

type Assessment struct {
	Verdict Verdict
	Counts  []RuleCount
	Hints   []string
}

func Assess(result analyzer.Result) Assessment {
	assessment := Assessment{Verdict: VerdictHealthy}
	counts := map[string]RuleCount{}

	for _, event := range result.Events {
		if event.Severity == analyzer.Danger {
			assessment.Verdict = VerdictDanger
		} else if event.Severity == analyzer.Warning && assessment.Verdict != VerdictDanger {
			assessment.Verdict = VerdictWarning
		}

		key := string(event.Severity) + "\x00" + event.Rule
		count := counts[key]
		count.Severity = event.Severity
		count.Rule = event.Rule
		count.Count++
		counts[key] = count
	}

	assessment.Counts = sortedCounts(counts)
	assessment.Hints = diagnosticHints(result)
	return assessment
}

func SummaryLine(assessment Assessment) string {
	switch assessment.Verdict {
	case VerdictDanger:
		return "DANGER: review attitude events before field use"
	case VerdictWarning:
		return "WARNING: inspect warnings and verify frame assumptions"
	default:
		return "HEALTHY: no attitude warnings detected"
	}
}

func FormatRuleCount(count RuleCount) string {
	return fmt.Sprintf("%s %s: %d events", count.Severity, count.Rule, count.Count)
}

func sortedCounts(counts map[string]RuleCount) []RuleCount {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	sorted := make([]RuleCount, 0, len(keys))
	for _, key := range keys {
		sorted = append(sorted, counts[key])
	}
	return sorted
}

func diagnosticHints(result analyzer.Result) []string {
	var hints []string

	if hasRule(result, "estimator_roll_mismatch") && result.MaxMismatchDeg >= 150 {
		hints = append(hints, "Near-180 deg roll mismatch: verify ENU/NED convention, quaternion ordering, and IMU-to-base_link transform before treating this as vehicle attitude.")
	}
	if hasRule(result, "yaw_rate") {
		hints = append(hints, "Yaw-rate warning: check timestamp spacing before diagnosing vehicle spin or thruster imbalance.")
	}
	if hasRule(result, "roll") || hasRule(result, "pitch") {
		hints = append(hints, "Body attitude warning: confirm IMU mounting calibration was applied for this vehicle.")
	}

	return unique(hints)
}

func hasRule(result analyzer.Result, rule string) bool {
	for _, event := range result.Events {
		if event.Rule == rule {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}
