# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-08-XX

### ⚠️ BREAKING CHANGES

- **Removed all legacy CLI flags**: The following flags are no longer supported:
  - `-b` (begin date) 
  - `-e` (end date)
  - `-s` (separator)
  - `-ifmt` (input format)
  - `-ofmt` (output format)
  - `-i` (interval)
  - `-tmpl` (template)
- **Expression syntax is now mandatory**: Must use `--expr` (or `-x`) parameter or provide expression via stdin
- **Default output format changed**: Now defaults to SQL format instead of custom format
- **Module path changed**: Now uses `github.com/sgaunet/calcdate/v2` for Go modules

### ✨ New Features

- **Modern expression syntax**: Natural language date expressions
  - Keywords: `today`, `tomorrow`, `yesterday`, `now`
  - Weekdays: `monday`, `tuesday`, etc. (next occurrence)
  - ISO dates: `2024-01-15`, `2024-01-15T14:30:00`
  - Relative dates: `+1d`, `-2w`, `+3M`, `-1Y`
  
- **Pipeline operations**: Chain multiple operations with `|`
  - Example: `today | +1M | endOfMonth` (last day of next month)
  - Example: `now | +2h | round hour` (2 hours from now, rounded to hour)
  
- **Range expressions with iterations**: Generate date ranges
  - Basic range: `today...+7d`
  - With iterations: `today...+7d --each=1d`
  - With transforms: `--transform '$begin +8h, $end +20h'`
  
- **Boundary operations**: 
  - Month: `startOfMonth`, `endOfMonth`
  - Week: `startOfWeek`, `endOfWeek`
  - Day: `startOfDay`, `endOfDay`
  - Hour: `startOfHour`, `endOfHour`
  - Minute: `startOfMinute`, `endOfMinute`
  - Second: `startOfSecond`, `endOfSecond`
  
- **Stdin support**: Pipe expressions directly
  - Example: `echo "today" | calcdate`
  - Example: `echo "today...+7d" | calcdate --each=1d`
  
- **Transform variables**: Use `$begin` and `$end` in iterations
  - Example: `--transform '$begin +9h, $end +17h'` for business hours
  
- **Skip weekends**: `--skip-weekends` flag for business day calculations

- **Additional time units**:
  - Business days: `+5bd` (5 business days)
  - Rounding: `round hour`, `round day`, etc.

### 📝 Improvements

- Complete rewrite of the date calculation engine
- New AST-based expression parser for better error handling
- More intuitive and readable syntax
- Better timezone handling with inline timezone parsing
- Extensive test coverage with unit and e2e tests
- Cleaner code architecture with separation of concerns

### 🔄 Migration from v1.x

See [MIGRATION.md](MIGRATION.md) for detailed instructions on upgrading from v1.x to v2.0.

### 📚 Documentation

- Updated README with comprehensive examples
- Added expression syntax quick reference
- Improved command-line help text
- Added stdin usage examples

## [1.5.1] - Previous release

- See GitHub releases for v1.x changelog