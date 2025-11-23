package calcdate

// OperationInfo contains metadata about a date operation.
type OperationInfo struct {
	Name        string   // Primary name of the operation
	Category    string   // Category for grouping (e.g., "Arithmetic", "Boundaries - Day")
	Description string   // Human-readable description
	Example     string   // Usage example
	Aliases     []string // Alternative names for this operation
}

// OperationCategory represents a group of related operations.
type OperationCategory struct {
	Name       string          // Category display name
	Operations []OperationInfo // Operations in this category
}

// GetOperationRegistry returns a comprehensive list of all available operations
// organized by category for documentation and help display purposes.
//
//nolint:funlen,maintidx // Function is long due to comprehensive operation data, not complexity
func GetOperationRegistry() []OperationCategory {
	return []OperationCategory{
		{
			Name: "Date Values",
			Operations: []OperationInfo{
				{
					Name:        "today",
					Category:    "Date Values",
					Description: "Start of today (00:00:00 in specified timezone)",
					Example:     `calcdate -x "today"`,
				},
				{
					Name:        "now",
					Category:    "Date Values",
					Description: "Current date and time",
					Example:     `calcdate -x "now"`,
				},
				{
					Name:        "yesterday",
					Category:    "Date Values",
					Description: "Start of yesterday (00:00:00)",
					Example:     `calcdate -x "yesterday"`,
				},
				{
					Name:        "tomorrow",
					Category:    "Date Values",
					Description: "Start of tomorrow (00:00:00)",
					Example:     `calcdate -x "tomorrow"`,
				},
				{
					Name:        "monday",
					Category:    "Date Values",
					Description: "Next Monday (00:00:00)",
					Example:     `calcdate -x "monday"`,
				},
				{
					Name:        "tuesday",
					Category:    "Date Values",
					Description: "Next Tuesday (00:00:00)",
					Example:     `calcdate -x "tuesday"`,
				},
				{
					Name:        "wednesday",
					Category:    "Date Values",
					Description: "Next Wednesday (00:00:00)",
					Example:     `calcdate -x "wednesday"`,
				},
				{
					Name:        "thursday",
					Category:    "Date Values",
					Description: "Next Thursday (00:00:00)",
					Example:     `calcdate -x "thursday"`,
				},
				{
					Name:        "friday",
					Category:    "Date Values",
					Description: "Next Friday (00:00:00)",
					Example:     `calcdate -x "friday"`,
				},
				{
					Name:        "saturday",
					Category:    "Date Values",
					Description: "Next Saturday (00:00:00)",
					Example:     `calcdate -x "saturday"`,
				},
				{
					Name:        "sunday",
					Category:    "Date Values",
					Description: "Next Sunday (00:00:00)",
					Example:     `calcdate -x "sunday"`,
				},
			},
		},
		{
			Name: "Arithmetic Operations",
			Operations: []OperationInfo{
				{
					Name:        "+<value><unit>",
					Category:    "Arithmetic Operations",
					Description: "Add time duration (e.g., +1d, +2h, +3M)",
					Example:     `calcdate -x "today +1d +2h"`,
				},
				{
					Name:        "-<value><unit>",
					Category:    "Arithmetic Operations",
					Description: "Subtract time duration (e.g., -1d, -2h, -3M)",
					Example:     `calcdate -x "now -3M -5d"`,
				},
			},
		},
		{
			Name: "Time Units",
			Operations: []OperationInfo{
				{
					Name:        "s",
					Category:    "Time Units",
					Description: "Seconds",
					Example:     `calcdate -x "now +30s"`,
				},
				{
					Name:        "m",
					Category:    "Time Units",
					Description: "Minutes",
					Example:     `calcdate -x "now +15m"`,
				},
				{
					Name:        "h",
					Category:    "Time Units",
					Description: "Hours",
					Example:     `calcdate -x "now +2h"`,
				},
				{
					Name:        "d",
					Category:    "Time Units",
					Description: "Days",
					Example:     `calcdate -x "today +7d"`,
				},
				{
					Name:        "w",
					Category:    "Time Units",
					Description: "Weeks (7 days)",
					Example:     `calcdate -x "today +2w"`,
				},
				{
					Name:        "M",
					Category:    "Time Units",
					Description: "Months",
					Example:     `calcdate -x "today +1M"`,
				},
				{
					Name:        "Y",
					Category:    "Time Units",
					Description: "Years",
					Example:     `calcdate -x "today +1Y"`,
				},
				{
					Name:        "q",
					Category:    "Time Units",
					Description: "Quarters (3 months)",
					Example:     `calcdate -x "today +1q"`,
				},
			},
		},
		{
			Name: "Boundaries - Day",
			Operations: []OperationInfo{
				{
					Name:        "start",
					Category:    "Boundaries - Day",
					Description: "Start of day (00:00:00)",
					Example:     `calcdate -x "now | start"`,
					Aliases:     []string{"startofday"},
				},
				{
					Name:        "startofday",
					Category:    "Boundaries - Day",
					Description: "Start of day (00:00:00)",
					Example:     `calcdate -x "now | startofday"`,
					Aliases:     []string{"start"},
				},
				{
					Name:        "end",
					Category:    "Boundaries - Day",
					Description: "End of day (23:59:59.999999999)",
					Example:     `calcdate -x "today | end"`,
					Aliases:     []string{"endofday"},
				},
				{
					Name:        "endofday",
					Category:    "Boundaries - Day",
					Description: "End of day (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofday"`,
					Aliases:     []string{"end"},
				},
			},
		},
		{
			Name: "Boundaries - Week",
			Operations: []OperationInfo{
				{
					Name:        "startofweek",
					Category:    "Boundaries - Week",
					Description: "Start of week (Monday 00:00:00)",
					Example:     `calcdate -x "today | startofweek"`,
				},
				{
					Name:        "endofweek",
					Category:    "Boundaries - Week",
					Description: "End of week (Sunday 23:59:59.999999999)",
					Example:     `calcdate -x "today | endofweek"`,
				},
			},
		},
		{
			Name: "Boundaries - Month",
			Operations: []OperationInfo{
				{
					Name:        "startofmonth",
					Category:    "Boundaries - Month",
					Description: "First day of month (00:00:00)",
					Example:     `calcdate -x "today | startofmonth"`,
				},
				{
					Name:        "endofmonth",
					Category:    "Boundaries - Month",
					Description: "Last day of month (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofmonth"`,
				},
			},
		},
		{
			Name: "Boundaries - Year",
			Operations: []OperationInfo{
				{
					Name:        "startofyear",
					Category:    "Boundaries - Year",
					Description: "January 1st (00:00:00)",
					Example:     `calcdate -x "today | startofyear"`,
				},
				{
					Name:        "endofyear",
					Category:    "Boundaries - Year",
					Description: "December 31st (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofyear"`,
				},
			},
		},
		{
			Name: "Boundaries - Quarter",
			Operations: []OperationInfo{
				{
					Name:        "startofquarter",
					Category:    "Boundaries - Quarter",
					Description: "First day of current quarter (00:00:00)",
					Example:     `calcdate -x "today | startofquarter"`,
				},
				{
					Name:        "endofquarter",
					Category:    "Boundaries - Quarter",
					Description: "Last day of current quarter (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofquarter"`,
				},
			},
		},
		{
			Name: "Boundaries - Time",
			Operations: []OperationInfo{
				{
					Name:        "startofhour",
					Category:    "Boundaries - Time",
					Description: "Start of current hour (:00:00)",
					Example:     `calcdate -x "now | startofhour"`,
				},
				{
					Name:        "endofhour",
					Category:    "Boundaries - Time",
					Description: "End of current hour (:59:59.999999999)",
					Example:     `calcdate -x "now | endofhour"`,
				},
				{
					Name:        "startofminute",
					Category:    "Boundaries - Time",
					Description: "Start of current minute (:00)",
					Example:     `calcdate -x "now | startofminute"`,
				},
				{
					Name:        "endofminute",
					Category:    "Boundaries - Time",
					Description: "End of current minute (:59.999999999)",
					Example:     `calcdate -x "now | endofminute"`,
				},
				{
					Name:        "startofsecond",
					Category:    "Boundaries - Time",
					Description: "Start of current second (.000000000)",
					Example:     `calcdate -x "now | startofsecond"`,
				},
				{
					Name:        "endofsecond",
					Category:    "Boundaries - Time",
					Description: "End of current second (.999999999)",
					Example:     `calcdate -x "now | endofsecond"`,
				},
			},
		},
		{
			Name: "Value Setters",
			Operations: []OperationInfo{
				{
					Name:        "day <value>",
					Category:    "Value Setters",
					Description: "Set specific day of month (1-31)",
					Example:     `calcdate -x "today | day 15"`,
				},
				{
					Name:        "time <HH:MM:SS>",
					Category:    "Value Setters",
					Description: "Set specific time",
					Example:     `calcdate -x "today | time 14:30:00"`,
				},
			},
		},
		{
			Name: "Transform Operations",
			Operations: []OperationInfo{
				{
					Name:        "round <unit>",
					Category:    "Transform Operations",
					Description: "Round to nearest unit (day, hour, minute)",
					Example:     `calcdate -x "now | round hour"`,
				},
				{
					Name:        "trunc <unit>",
					Category:    "Transform Operations",
					Description: "Truncate to unit boundary (day, hour, minute)",
					Example:     `calcdate -x "now | trunc hour"`,
				},
			},
		},
		{
			Name: "Range Operations",
			Operations: []OperationInfo{
				{
					Name:        "...",
					Category:    "Range Operations",
					Description: "Create date range (requires -each flag for iteration)",
					Example:     `calcdate -x "today...+7d" -each 1d`,
				},
			},
		},
		{
			Name: "Pipeline Operations",
			Operations: []OperationInfo{
				{
					Name:        "|",
					Category:    "Pipeline Operations",
					Description: "Chain operations together",
					Example:     `calcdate -x "today | +1M | endofmonth"`,
				},
			},
		},
	}
}
