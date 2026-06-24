package cli

import (
	"fmt"
	"io"
	"os"

	"blue-horizon/internal/analyzer"
	"blue-horizon/internal/calibration"
	"blue-horizon/internal/config"
	"blue-horizon/internal/input"
	"blue-horizon/internal/report"

	"github.com/spf13/cobra"
)

type options struct {
	configPath      string
	calibrationPath string
	outputPath      string
	format          string
}

func NewRootCommand() *cobra.Command {
	opts := &options{}

	root := &cobra.Command{
		Use:   "blue-horizon",
		Short: "Analyze AUV attitude logs for unsafe attitude and estimator mismatch",
		Long: `Blue Horizon analyzes exported AUV attitude CSV logs.

It checks roll, pitch, yaw rate, and IMU/GTSAM estimator disagreement while
printing the frame assumptions that must be verified before trusting warnings.`,
		SilenceUsage: true,
	}

	root.AddCommand(newAnalyzeCommand(opts, os.Stdout))
	root.AddCommand(newReportCommand(opts, os.Stdout))
	root.AddCommand(newCalibrateCommand(opts, os.Stdout))
	return root
}

func newAnalyzeCommand(opts *options, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze data.csv",
		Short: "Analyze a CSV attitude log and print warnings plus a summary report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(args[0], opts, out, false)
		},
	}
	cmd.Flags().StringVar(&opts.configPath, "config", "", "config YAML path")
	cmd.Flags().StringVar(&opts.calibrationPath, "calibration", "", "calibration YAML path")
	cmd.Flags().StringVar(&opts.format, "format", "text", "report format: text or markdown")
	return cmd
}

func newReportCommand(opts *options, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report data.csv",
		Short: "Analyze a CSV attitude log and print only the report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(args[0], opts, out, true)
		},
	}
	cmd.Flags().StringVar(&opts.configPath, "config", "", "config YAML path")
	cmd.Flags().StringVar(&opts.calibrationPath, "calibration", "", "calibration YAML path")
	cmd.Flags().StringVar(&opts.format, "format", "text", "report format: text or markdown")
	return cmd
}

func newCalibrateCommand(opts *options, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calibrate still_sample.csv",
		Short: "Estimate static IMU mounting offsets from a still, level sample",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCalibrate(args[0], opts, out)
		},
	}
	cmd.Flags().StringVar(&opts.outputPath, "output", "calibration.yaml", "output calibration YAML path")
	return cmd
}

func runAnalyze(csvPath string, opts *options, out io.Writer, reportOnly bool) error {
	printFrameAssumptions(out)

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return err
	}
	cal, err := calibration.Load(opts.calibrationPath)
	if err != nil {
		return err
	}
	samples, err := input.ReadCSV(csvPath)
	if err != nil {
		return err
	}

	result := analyzer.Analyze(samples, cfg, cal)
	if !reportOnly {
		for _, event := range result.Events {
			fmt.Fprintln(out, analyzer.FormatEvent(event))
		}
		fmt.Fprintln(out)
	}

	switch opts.format {
	case "text":
		report.PrintText(out, result)
	case "markdown":
		report.PrintMarkdown(out, result)
	default:
		return fmt.Errorf("unknown format %q", opts.format)
	}

	return nil
}

func runCalibrate(csvPath string, opts *options, out io.Writer) error {
	printFrameAssumptions(out)

	samples, err := input.ReadCSV(csvPath)
	if err != nil {
		return err
	}
	cal, err := calibration.Estimate(samples)
	if err != nil {
		return err
	}
	if err := calibration.Save(opts.outputPath, cal); err != nil {
		return err
	}

	fmt.Fprintln(out, "Estimated IMU mounting offset:")
	fmt.Fprintf(out, "roll_offset_deg: %.3f\n", cal.RollOffsetDeg)
	fmt.Fprintf(out, "pitch_offset_deg: %.3f\n", cal.PitchOffsetDeg)
	fmt.Fprintf(out, "Saved calibration to %s\n", opts.outputPath)
	return nil
}

func printFrameAssumptions(out io.Writer) {
	fmt.Fprintln(out, "Frame assumptions: base_link x forward, y left, z up; roll around x, pitch around y, yaw around z.")
	fmt.Fprintln(out, "Verify ENU/NED/body-frame conventions before trusting warnings.")
	fmt.Fprintln(out)
}
