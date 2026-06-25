package calcdate

// Date keyword constants.
const (
	todayKeyword     = "today"
	nowKeyword       = "now"
	yesterdayKeyword = "yesterday"
	tomorrowKeyword  = "tomorrow"
	mondayKeyword    = "monday"
	tuesdayKeyword   = "tuesday"
	wednesdayKeyword = "wednesday"
	thursdayKeyword  = "thursday"
	fridayKeyword    = "friday"
	saturdayKeyword  = "saturday"
	sundayKeyword    = "sunday"
)

// Boundary operation keyword constants.
const (
	startKeyword          = "start"
	endKeyword            = "end"
	startOfDayKeyword     = "startofday"
	endOfDayKeyword       = "endofday"
	startOfWeekKeyword    = "startofweek"
	endOfWeekKeyword      = "endofweek"
	startOfMonthKeyword   = "startofmonth"
	endOfMonthKeyword     = "endofmonth"
	startOfYearKeyword    = "startofyear"
	endOfYearKeyword      = "endofyear"
	startOfQuarterKeyword = "startofquarter"
	endOfQuarterKeyword   = "endofquarter"
	startOfHourKeyword    = "startofhour"
	endOfHourKeyword      = "endofhour"
	startOfMinuteKeyword  = "startofminute"
	endOfMinuteKeyword    = "endofminute"
	startOfSecondKeyword  = "startofsecond"
	endOfSecondKeyword    = "endofsecond"
)

// Time unit and transform keyword constants.
const (
	dayKeyword    = "day"
	hourKeyword   = "hour"
	minuteKeyword = "minute"
	roundKeyword  = "round"
	truncKeyword  = "trunc"
)

// Range operator constant.
const rangeOperator = "..."

// Operation category name constants.
const (
	categoryDateValues        = "Date Values"
	categoryArithmetic        = "Arithmetic Operations"
	categoryTimeUnits         = "Time Units"
	categoryBoundariesDay     = "Boundaries - Day"
	categoryBoundariesWeek    = "Boundaries - Week"
	categoryBoundariesMonth   = "Boundaries - Month"
	categoryBoundariesYear    = "Boundaries - Year"
	categoryBoundariesQuarter = "Boundaries - Quarter"
	categoryBoundariesTime    = "Boundaries - Time"
	categoryValueSetters      = "Value Setters"
	categoryTransform         = "Transform Operations"
)

// Unix filesystem path constants.
const unixZoneinfoPath = "/usr/share/zoneinfo/"
