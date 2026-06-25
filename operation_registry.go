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
			Name: categoryDateValues,
			Operations: []OperationInfo{
				{
					Name:        todayKeyword,
					Category:    categoryDateValues,
					Description: "Start of today (00:00:00 in specified timezone)",
					Example:     `calcdate -x "today"`,
				},
				{
					Name:        nowKeyword,
					Category:    categoryDateValues,
					Description: "Current date and time",
					Example:     `calcdate -x "now"`,
				},
				{
					Name:        yesterdayKeyword,
					Category:    categoryDateValues,
					Description: "Start of yesterday (00:00:00)",
					Example:     `calcdate -x "yesterday"`,
				},
				{
					Name:        tomorrowKeyword,
					Category:    categoryDateValues,
					Description: "Start of tomorrow (00:00:00)",
					Example:     `calcdate -x "tomorrow"`,
				},
				{
					Name:        mondayKeyword,
					Category:    categoryDateValues,
					Description: "Next Monday (00:00:00)",
					Example:     `calcdate -x "monday"`,
				},
				{
					Name:        tuesdayKeyword,
					Category:    categoryDateValues,
					Description: "Next Tuesday (00:00:00)",
					Example:     `calcdate -x "tuesday"`,
				},
				{
					Name:        wednesdayKeyword,
					Category:    categoryDateValues,
					Description: "Next Wednesday (00:00:00)",
					Example:     `calcdate -x "wednesday"`,
				},
				{
					Name:        thursdayKeyword,
					Category:    categoryDateValues,
					Description: "Next Thursday (00:00:00)",
					Example:     `calcdate -x "thursday"`,
				},
				{
					Name:        fridayKeyword,
					Category:    categoryDateValues,
					Description: "Next Friday (00:00:00)",
					Example:     `calcdate -x "friday"`,
				},
				{
					Name:        saturdayKeyword,
					Category:    categoryDateValues,
					Description: "Next Saturday (00:00:00)",
					Example:     `calcdate -x "saturday"`,
				},
				{
					Name:        sundayKeyword,
					Category:    categoryDateValues,
					Description: "Next Sunday (00:00:00)",
					Example:     `calcdate -x "sunday"`,
				},
			},
		},
		{
			Name: categoryArithmetic,
			Operations: []OperationInfo{
				{
					Name:        "+<value><unit>",
					Category:    categoryArithmetic,
					Description: "Add time duration (e.g., +1d, +2h, +3M)",
					Example:     `calcdate -x "today +1d +2h"`,
				},
				{
					Name:        "-<value><unit>",
					Category:    categoryArithmetic,
					Description: "Subtract time duration (e.g., -1d, -2h, -3M)",
					Example:     `calcdate -x "now -3M -5d"`,
				},
			},
		},
		{
			Name: categoryTimeUnits,
			Operations: []OperationInfo{
				{
					Name:        "s",
					Category:    categoryTimeUnits,
					Description: "Seconds",
					Example:     `calcdate -x "now +30s"`,
				},
				{
					Name:        "m",
					Category:    categoryTimeUnits,
					Description: "Minutes",
					Example:     `calcdate -x "now +15m"`,
				},
				{
					Name:        "h",
					Category:    categoryTimeUnits,
					Description: "Hours",
					Example:     `calcdate -x "now +2h"`,
				},
				{
					Name:        "d",
					Category:    categoryTimeUnits,
					Description: "Days",
					Example:     `calcdate -x "today +7d"`,
				},
				{
					Name:        "w",
					Category:    categoryTimeUnits,
					Description: "Weeks (7 days)",
					Example:     `calcdate -x "today +2w"`,
				},
				{
					Name:        "M",
					Category:    categoryTimeUnits,
					Description: "Months",
					Example:     `calcdate -x "today +1M"`,
				},
				{
					Name:        "Y",
					Category:    categoryTimeUnits,
					Description: "Years",
					Example:     `calcdate -x "today +1Y"`,
				},
				{
					Name:        "q",
					Category:    categoryTimeUnits,
					Description: "Quarters (3 months)",
					Example:     `calcdate -x "today +1q"`,
				},
			},
		},
		{
			Name: categoryBoundariesDay,
			Operations: []OperationInfo{
				{
					Name:        startKeyword,
					Category:    categoryBoundariesDay,
					Description: "Start of day (00:00:00)",
					Example:     `calcdate -x "now | start"`,
					Aliases:     []string{startOfDayKeyword},
				},
				{
					Name:        startOfDayKeyword,
					Category:    categoryBoundariesDay,
					Description: "Start of day (00:00:00)",
					Example:     `calcdate -x "now | startofday"`,
					Aliases:     []string{startKeyword},
				},
				{
					Name:        endKeyword,
					Category:    categoryBoundariesDay,
					Description: "End of day (23:59:59.999999999)",
					Example:     `calcdate -x "today | end"`,
					Aliases:     []string{endOfDayKeyword},
				},
				{
					Name:        endOfDayKeyword,
					Category:    categoryBoundariesDay,
					Description: "End of day (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofday"`,
					Aliases:     []string{endKeyword},
				},
			},
		},
		{
			Name: categoryBoundariesWeek,
			Operations: []OperationInfo{
				{
					Name:        startOfWeekKeyword,
					Category:    categoryBoundariesWeek,
					Description: "Start of week (Monday 00:00:00)",
					Example:     `calcdate -x "today | startofweek"`,
				},
				{
					Name:        endOfWeekKeyword,
					Category:    categoryBoundariesWeek,
					Description: "End of week (Sunday 23:59:59.999999999)",
					Example:     `calcdate -x "today | endofweek"`,
				},
			},
		},
		{
			Name: categoryBoundariesMonth,
			Operations: []OperationInfo{
				{
					Name:        startOfMonthKeyword,
					Category:    categoryBoundariesMonth,
					Description: "First day of month (00:00:00)",
					Example:     `calcdate -x "today | startofmonth"`,
				},
				{
					Name:        endOfMonthKeyword,
					Category:    categoryBoundariesMonth,
					Description: "Last day of month (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofmonth"`,
				},
			},
		},
		{
			Name: categoryBoundariesYear,
			Operations: []OperationInfo{
				{
					Name:        startOfYearKeyword,
					Category:    categoryBoundariesYear,
					Description: "January 1st (00:00:00)",
					Example:     `calcdate -x "today | startofyear"`,
				},
				{
					Name:        endOfYearKeyword,
					Category:    categoryBoundariesYear,
					Description: "December 31st (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofyear"`,
				},
			},
		},
		{
			Name: categoryBoundariesQuarter,
			Operations: []OperationInfo{
				{
					Name:        startOfQuarterKeyword,
					Category:    categoryBoundariesQuarter,
					Description: "First day of current quarter (00:00:00)",
					Example:     `calcdate -x "today | startofquarter"`,
				},
				{
					Name:        endOfQuarterKeyword,
					Category:    categoryBoundariesQuarter,
					Description: "Last day of current quarter (23:59:59.999999999)",
					Example:     `calcdate -x "today | endofquarter"`,
				},
			},
		},
		{
			Name: categoryBoundariesTime,
			Operations: []OperationInfo{
				{
					Name:        startOfHourKeyword,
					Category:    categoryBoundariesTime,
					Description: "Start of current hour (:00:00)",
					Example:     `calcdate -x "now | startofhour"`,
				},
				{
					Name:        endOfHourKeyword,
					Category:    categoryBoundariesTime,
					Description: "End of current hour (:59:59.999999999)",
					Example:     `calcdate -x "now | endofhour"`,
				},
				{
					Name:        startOfMinuteKeyword,
					Category:    categoryBoundariesTime,
					Description: "Start of current minute (:00)",
					Example:     `calcdate -x "now | startofminute"`,
				},
				{
					Name:        endOfMinuteKeyword,
					Category:    categoryBoundariesTime,
					Description: "End of current minute (:59.999999999)",
					Example:     `calcdate -x "now | endofminute"`,
				},
				{
					Name:        startOfSecondKeyword,
					Category:    categoryBoundariesTime,
					Description: "Start of current second (.000000000)",
					Example:     `calcdate -x "now | startofsecond"`,
				},
				{
					Name:        endOfSecondKeyword,
					Category:    categoryBoundariesTime,
					Description: "End of current second (.999999999)",
					Example:     `calcdate -x "now | endofsecond"`,
				},
			},
		},
		{
			Name: categoryValueSetters,
			Operations: []OperationInfo{
				{
					Name:        "day <value>",
					Category:    categoryValueSetters,
					Description: "Set specific day of month (1-31)",
					Example:     `calcdate -x "today | day 15"`,
				},
				{
					Name:        "time <HH:MM:SS>",
					Category:    categoryValueSetters,
					Description: "Set specific time",
					Example:     `calcdate -x "today | time 14:30:00"`,
				},
			},
		},
		{
			Name: categoryTransform,
			Operations: []OperationInfo{
				{
					Name:        "round <unit>",
					Category:    categoryTransform,
					Description: "Round to nearest unit (day, hour, minute)",
					Example:     `calcdate -x "now | round hour"`,
				},
				{
					Name:        "trunc <unit>",
					Category:    categoryTransform,
					Description: "Truncate to unit boundary (day, hour, minute)",
					Example:     `calcdate -x "now | trunc hour"`,
				},
			},
		},
		{
			Name: "Range Operations",
			Operations: []OperationInfo{
				{
					Name:        rangeOperator,
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
