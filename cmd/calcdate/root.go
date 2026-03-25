package main

import (
	"fmt"
	"os"

	calcdate "github.com/sgaunet/calcdate/v2"
	"github.com/spf13/cobra"
)

var version = "development"

var (
	flagExpr         string
	flagEach         string
	flagTransform    string
	flagFormat       string
	flagTZ           string
	flagListOps      bool
	flagListTZ       bool
	flagVersion      bool
	flagSkipWeekends bool
)

var rootCmd = &cobra.Command{
	Use:   "calcdate [flags]",
	Short: "A modern command-line utility for date calculations",
	Long: `calcdate is a modern command-line utility for date calculations and operations.
It's particularly useful for generating date ranges for database queries,
automating maintenance windows, and performing complex date arithmetic
with an intuitive expression syntax.`,
	Example: `  calcdate -x "today +1d"                    # Tomorrow
  calcdate -x "now | +2h | round hour"       # 2 hours from now, rounded
  calcdate -x "today | startofmonth"         # First day of month
  calcdate -x "today...+7d" --each 1d         # Next 7 days

Use --list-ops to see all available operations`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runRoot,
}

func init() {
	rootCmd.Flags().StringVarP(&flagExpr, "expr", "x", "",
		"Date expression (e.g., 'today +1d', 'now | +2h | round hour', 'today...+7d')")
	rootCmd.Flags().StringVar(&flagEach, "each", "",
		"Iteration interval for ranges (e.g., '1d', '1w', '1M')")
	rootCmd.Flags().StringVarP(&flagTransform, "transform", "t", "",
		"Transform expression for iterations (e.g., '$begin +8h, $end +20h')")
	rootCmd.Flags().StringVarP(&flagFormat, "format", "f", "",
		"Output format: iso, sql, ts, human, compact, or Unix date format "+
			"(e.g., '%%Y-%%m-%%d %%H:%%M:%%S')")
	rootCmd.Flags().StringVar(&flagTZ, "tz", "Local", "Input timezone")
	rootCmd.Flags().BoolVar(&flagListOps, "list-ops", false,
		"List all available operations")
	rootCmd.Flags().BoolVar(&flagListTZ, "list-tz", false,
		"List timezones")
	rootCmd.Flags().BoolVarP(&flagVersion, "version", "v", false,
		"Print version")
	rootCmd.Flags().BoolVar(&flagSkipWeekends, "skip-weekends", false,
		"Skip weekend days in iterations")

	registerCompletions()
}

func runRoot(_ *cobra.Command, _ []string) error {
	if flagListTZ {
		calcdate.ListTZ()
		return nil
	}

	if flagListOps {
		printOperationsList()
		return nil
	}

	if flagVersion {
		printVersion()
		return nil
	}

	expr := flagExpr
	if expr == "" {
		if isStdinRedirected() {
			var err error
			expr, err = readExprFromStdin()
			if err != nil {
				return fmt.Errorf("error reading from stdin: %w", err)
			}
		} else {
			expr = "now"
		}
	}

	processExpressionMode(expr, flagEach, flagTransform, flagFormat, flagTZ, flagSkipWeekends)
	return nil
}

func printVersion() {
	fmt.Println(version)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
