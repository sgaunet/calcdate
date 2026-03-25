package main

import (
	"os"
	"strings"

	calcdate "github.com/sgaunet/calcdate/v2"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for calcdate.

To load completions:

Bash:
  $ source <(calcdate completion bash)
  # To persist, add to ~/.bashrc or install:
  $ calcdate completion bash > /etc/bash_completion.d/calcdate

Zsh:
  $ source <(calcdate completion zsh)
  # To persist:
  $ calcdate completion zsh > "${fpath[1]}/_calcdate"

Fish:
  $ calcdate completion fish | source
  # To persist:
  $ calcdate completion fish > ~/.config/fish/completions/calcdate.fish

PowerShell:
  PS> calcdate completion powershell | Out-String | Invoke-Expression
  # To persist, add the output to your PowerShell profile
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(_ *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func registerCompletions() {
	_ = rootCmd.RegisterFlagCompletionFunc("format", completeFormat)
	_ = rootCmd.RegisterFlagCompletionFunc("tz", completeTimezone)
	_ = rootCmd.RegisterFlagCompletionFunc("expr", completeExpression)
	_ = rootCmd.RegisterFlagCompletionFunc("each", completeInterval)
	_ = rootCmd.RegisterFlagCompletionFunc("transform", completeTransform)
}

func completeFormat(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	formats := []string{
		"iso\tISO 8601 format (2024-01-15T00:00:00Z)",
		"sql\tSQL format (2024-01-15 00:00:00)",
		"ts\tUnix timestamp",
		"human\tHuman readable (Monday, January 15, 2024)",
		"compact\tCompact format (20240115)",
	}
	return formats, cobra.ShellCompDirectiveNoFileComp
}

func completeTimezone(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	timezones := calcdate.GetTimezones()
	if toComplete == "" {
		return timezones, cobra.ShellCompDirectiveNoFileComp
	}
	var matches []string
	lower := strings.ToLower(toComplete)
	for _, tz := range timezones {
		if strings.HasPrefix(strings.ToLower(tz), lower) {
			matches = append(matches, tz)
		}
	}
	if len(matches) == 0 {
		return timezones, cobra.ShellCompDirectiveNoFileComp
	}
	return matches, cobra.ShellCompDirectiveNoFileComp
}

func completeExpression(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	keywords := []string{
		"today\tStart of today (00:00:00)",
		"now\tCurrent date and time",
		"yesterday\tStart of yesterday",
		"tomorrow\tStart of tomorrow",
		"monday\tNext Monday",
		"tuesday\tNext Tuesday",
		"wednesday\tNext Wednesday",
		"thursday\tNext Thursday",
		"friday\tNext Friday",
		"saturday\tNext Saturday",
		"sunday\tNext Sunday",
	}
	return keywords, cobra.ShellCompDirectiveNoFileComp
}

func completeInterval(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	intervals := []string{
		"1d\tDaily",
		"1w\tWeekly",
		"1M\tMonthly",
		"1q\tQuarterly",
		"1Y\tYearly",
		"1h\tHourly",
		"30m\tEvery 30 minutes",
		"15m\tEvery 15 minutes",
	}
	return intervals, cobra.ShellCompDirectiveNoFileComp
}

func completeTransform(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	transforms := []string{
		"$begin +8h, $end +20h\tBusiness hours (8am-8pm)",
		"$begin +9h, $end +17h\tOffice hours (9am-5pm)",
		"$begin, $end -1s\tExclusive end",
	}
	return transforms, cobra.ShellCompDirectiveNoFileComp
}
