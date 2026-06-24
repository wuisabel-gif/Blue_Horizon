package report

import (
	"fmt"
	"io"
	"sort"

	"blue-horizon/internal/analyzer"
)

func PrintText(w io.Writer, result analyzer.Result) {
	fmt.Fprintln(w, "Blue Horizon Report")
	fmt.Fprintf(w, "Samples processed: %d\n", result.SamplesProcessed)
	fmt.Fprintf(w, "Duration: %.2f s\n", result.DurationSec)
	fmt.Fprintf(w, "Max roll: %.1f deg\n", result.MaxRollDeg)
	fmt.Fprintf(w, "Max pitch: %.1f deg\n", result.MaxPitchDeg)
	fmt.Fprintf(w, "Max IMU/GTSAM mismatch: %.1f deg\n", result.MaxMismatchDeg)
	fmt.Fprintln(w, "Warnings:")
	for _, line := range eventCounts(result) {
		fmt.Fprintf(w, "- %s\n", line)
	}
	if result.WorstEvent != nil {
		fmt.Fprintln(w, "Worst event:")
		fmt.Fprintf(w, "t=%.2f s\n", result.WorstEvent.TimeSec)
		fmt.Fprintf(w, "%s=%.1f deg\n", result.WorstEvent.Rule, result.WorstEvent.ValueDeg)
	}
}

func PrintMarkdown(w io.Writer, result analyzer.Result) {
	fmt.Fprintln(w, "# Blue Horizon Report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- Samples processed: %d\n", result.SamplesProcessed)
	fmt.Fprintf(w, "- Duration: %.2f s\n", result.DurationSec)
	fmt.Fprintf(w, "- Max roll: %.1f deg\n", result.MaxRollDeg)
	fmt.Fprintf(w, "- Max pitch: %.1f deg\n", result.MaxPitchDeg)
	fmt.Fprintf(w, "- Max IMU/GTSAM mismatch: %.1f deg\n", result.MaxMismatchDeg)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Warnings")
	for _, line := range eventCounts(result) {
		fmt.Fprintf(w, "- %s\n", line)
	}
	if result.WorstEvent != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## Worst Event")
		fmt.Fprintf(w, "- t=%.2f s\n", result.WorstEvent.TimeSec)
		fmt.Fprintf(w, "- %s: %.1f deg\n", result.WorstEvent.Rule, result.WorstEvent.ValueDeg)
	}
}

func eventCounts(result analyzer.Result) []string {
	counts := map[string]int{}
	for _, event := range result.Events {
		counts[string(event.Severity)+" "+event.Rule]++
	}
	if len(counts) == 0 {
		return []string{"None"}
	}

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %d events", key, counts[key]))
	}
	return lines
}
