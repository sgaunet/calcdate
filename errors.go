package calcdate

import "errors"

// Static errors for consistent error handling.
var (
	// Parser errors.
	ErrOperationWithoutBaseDate    = errors.New("operation node cannot be evaluated without a base date")
	ErrInvalidPipelineOperation    = errors.New("invalid operation in pipeline")
	ErrRangeNodesSeparateHandling  = errors.New("range nodes must be handled separately")
	ErrVariableNotFound            = errors.New("variable not found in context")
	ErrTransformNodesSeparate      = errors.New("transform nodes must be handled separately")
	ErrNotRangeExpression          = errors.New("not a range expression")
	ErrTransformPartsInvalid       = errors.New("transform must have exactly two parts separated by comma")
	ErrUnexpectedEndOfExpression   = errors.New("unexpected end of expression")
	ErrUnexpectedToken             = errors.New("unexpected token")
	ErrExpectedUnitAfterOperator   = errors.New("expected unit or number after operator")
	ErrExpectedOperationAfterPipe  = errors.New("expected operation after pipe")
	ErrExpectedDateAfterOperator   = errors.New("expected date after operator")
	ErrUnsupportedTokenPrimary     = errors.New("unsupported token type in primary expression")
	ErrUnknownUnit                 = errors.New("unknown unit")
	ErrInvalidDateValue            = errors.New("invalid date value")
	ErrUnknownOperation            = errors.New("unknown operation")
	ErrInvalidInterval             = errors.New("invalid interval format")
	ErrMissingOperand              = errors.New("missing operand after operator")
	ErrInvalidTransformExpression  = errors.New("invalid transform expression")
	ErrTooManyIterations           = errors.New("too many iterations (max 10000)")
	ErrInvalidTimezone             = errors.New("invalid timezone")
	ErrInvalidIntervalWithoutEnd   = errors.New("interval without end date")
	ErrInvalidInput                = errors.New("invalid input")
	ErrUnexpectedCharacter         = errors.New("unexpected character")
	ErrEmptyExpression             = errors.New("empty expression")
	ErrInvalidNumberFormat         = errors.New("invalid number format")
	ErrInvalidISODateFormat        = errors.New("invalid ISO date format")
	ErrInvalidKeyword              = errors.New("invalid keyword")
	ErrInvalidVariable             = errors.New("invalid variable")
	ErrInvalidOperator             = errors.New("invalid operator")
	ErrInvalidUnit                 = errors.New("invalid unit")
	ErrInvalidRange                = errors.New("invalid range")
	ErrInvalidPipe                 = errors.New("invalid pipe")
	ErrInvalidTransform            = errors.New("invalid transform")
	ErrInvalidComma                = errors.New("invalid comma")
	ErrInvalidParenthesis          = errors.New("invalid parenthesis")
	ErrInvalidFlag                 = errors.New("invalid flag")
	ErrInvalidDate                 = errors.New("invalid date")
	ErrInvalidTime                 = errors.New("invalid time")
	ErrInvalidDuration             = errors.New("invalid duration")
	ErrInvalidQuarter              = errors.New("invalid quarter")
	ErrInvalidWeekday              = errors.New("invalid weekday")
	ErrInvalidMonth                = errors.New("invalid month")
	ErrInvalidYear                 = errors.New("invalid year")
	ErrInvalidDay                  = errors.New("invalid day")
	ErrInvalidHour                 = errors.New("invalid hour")
	ErrInvalidMinute               = errors.New("invalid minute")
	ErrInvalidSecond               = errors.New("invalid second")
)

// Constants for magic numbers.
const (
	// Time-related constants.
	HoursInDay       = 24
	MinutesInHour    = 60
	SecondsInMinute  = 60
	DaysInWeek       = 7
	MonthsInYear     = 12
	MonthsInQuarter  = 3
	QuartersInYear   = 4
	MaxIterations    = 10000
	
	// Date boundaries.
	FirstQuarterEnd  = 3
	SecondQuarterEnd = 6
	ThirdQuarterEnd  = 9
	FourthQuarterEnd = 12
	
	// Time boundaries.
	NoonHour         = 12
	HalfMinute       = 30
	HalfSecond       = 30
	
	// Parser constants.
	MinIntervalLength = 2
	TransformParts    = 2
	
	// Complexity thresholds.
	MaxTokenizerNestDepth = 10
)