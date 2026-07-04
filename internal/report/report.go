package report

import (
	"fmt"
	"io"

	"blue-horizon/internal/analyzer"
	"blue-horizon/internal/health"
)

func PrintText(w io.Writer, result analyzer.Result) {
	assessment := health.Assess(result)

	fmt.Fprintln(w, "Blue Horizon Report")
	fmt.Fprintf(w, "Verdict: %s\n", assessment.Verdict)
	fmt.Fprintf(w, "Summary: %s\n", health.SummaryLine(assessment))
	fmt.Fprintf(w, "Samples processed: %d\n", result.SamplesProcessed)
	fmt.Fprintf(w, "Duration: %.2f s\n", result.DurationSec)
	fmt.Fprintf(w, "Max roll: %.1f deg\n", result.MaxRollDeg)
	fmt.Fprintf(w, "Max pitch: %.1f deg\n", result.MaxPitchDeg)
	fmt.Fprintf(w, "Max IMU/GTSAM mismatch: %.1f deg\n", result.MaxMismatchDeg)
	fmt.Fprintln(w, "Event counts:")
	for _, line := range eventCountLines(assessment) {
		fmt.Fprintf(w, "- %s\n", line)
	}
	if result.WorstEvent != nil {
		fmt.Fprintln(w, "Worst event:")
		fmt.Fprintf(w, "t=%.2f s\n", result.WorstEvent.TimeSec)
		fmt.Fprintf(w, "%s=%.1f %s\n", result.WorstEvent.Rule, result.WorstEvent.ValueDeg, analyzer.Unit(result.WorstEvent.Rule))
	}
	if len(assessment.Hints) > 0 {
		fmt.Fprintln(w, "Diagnostic hints:")
		for _, hint := range assessment.Hints {
			fmt.Fprintf(w, "- %s\n", hint)
		}
	}
}

func PrintMarkdown(w io.Writer, result analyzer.Result) {
	assessment := health.Assess(result)

	fmt.Fprintln(w, "# Blue Horizon Report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- Verdict: **%s**\n", assessment.Verdict)
	fmt.Fprintf(w, "- Summary: %s\n", health.SummaryLine(assessment))
	fmt.Fprintf(w, "- Samples processed: %d\n", result.SamplesProcessed)
	fmt.Fprintf(w, "- Duration: %.2f s\n", result.DurationSec)
	fmt.Fprintf(w, "- Max roll: %.1f deg\n", result.MaxRollDeg)
	fmt.Fprintf(w, "- Max pitch: %.1f deg\n", result.MaxPitchDeg)
	fmt.Fprintf(w, "- Max IMU/GTSAM mismatch: %.1f deg\n", result.MaxMismatchDeg)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Event Counts")
	for _, line := range eventCountLines(assessment) {
		fmt.Fprintf(w, "- %s\n", line)
	}
	if result.WorstEvent != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## Worst Event")
		fmt.Fprintf(w, "- t=%.2f s\n", result.WorstEvent.TimeSec)
		fmt.Fprintf(w, "- %s: %.1f %s\n", result.WorstEvent.Rule, result.WorstEvent.ValueDeg, analyzer.Unit(result.WorstEvent.Rule))
	}
	if len(assessment.Hints) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## Diagnostic Hints")
		for _, hint := range assessment.Hints {
			fmt.Fprintf(w, "- %s\n", hint)
		}
	}
}

func eventCountLines(assessment health.Assessment) []string {
	if len(assessment.Counts) == 0 {
		return []string{"None"}
	}

	lines := make([]string, 0, len(assessment.Counts))
	for _, count := range assessment.Counts {
		lines = append(lines, health.FormatRuleCount(count))
	}
	return lines
}
