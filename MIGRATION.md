# Migration Guide: v1.x to v2.0

This guide helps you migrate from calcdate v1.x to v2.0. Version 2.0 introduces a completely new expression syntax and removes the legacy command-line flags.

## ⚠️ Breaking Changes Summary

1. **All legacy flags removed** (`-b`, `-e`, `-s`, `-ifmt`, `-ofmt`, `-i`, `-tmpl`)
2. **Expression syntax (`--expr` or `-x`) is now required**
3. **Default output format changed to SQL**
4. **Go module path changed to `github.com/sgaunet/calcdate/v2`**

## Installation

### For Go Users
```bash
# Update your go.mod
go get github.com/sgaunet/calcdate/v2@latest
```

### For Homebrew Users
```bash
brew upgrade calcdate
```

## Command Translation Guide

### Basic Date Calculations

#### v1.x: Display current date
```bash
calcdate -b "// ::"
```

or 
```bash
calcdate
```

#### v2.0: Display current date
```bash
calcdate --expr "now"
```
or 
```bash
calcdate
```

---

### Date Arithmetic

#### v1.x: Add days to current date
```bash
calcdate -b "// ::" -e "//+7 ::"
```

#### v2.0: Add days to current date

with 2 commands:
```bash
calcdate --expr 'today | startOfDay'
calcdate --expr 'today+7d | endOfDay'
```

---

### Date Ranges

#### v1.x: Create date range
```bash
calcdate -b "// ::" -e "//+30 ::" -i 24h -tmpl "{{ .BeginTime }} - {{ .EndTime }}"
```

#### v2.0: Create date range
```bash
calcdate --expr "today...+30d" --each=1d --format="%Y-%m-%d %H:%M:%S %Z"
```

---

### Custom Formats

#### v1.x: Custom output format
```bash
calcdate -b "// ::" -ofmt "%YYYY-%MM-%DD %hh:%mm:%ss"
```

#### v2.0: Custom output format
```bash
# Using predefined formats
calcdate --expr "today" --format=iso    # ISO 8601
calcdate --expr "today" --format=sql    # SQL datetime
calcdate --expr "today" --format=ts     # Unix timestamp
calcdate --expr "today" --format=human  # Human readable
calcdate --expr "today" --format=compact # Compact format

# Using Unix date format
calcdate --expr "today" --format="%Y-%m-%d %H:%M:%S"
```

---

### Complex Examples

#### v1.x: Business hours for next week
```bash
# Not easily achievable in v1.x
```

#### v2.0: Business hours for next week
```bash
calcdate --expr "today...+7d" \
  --each=1d \
  --skip-weekends \
  --transform='$begin +9h, $end +17h'
```

---

## New Features in v2.0

### Natural Language Dates
```bash
calcdate --expr "today"
calcdate --expr "tomorrow"
calcdate --expr "yesterday"
calcdate --expr "monday"    # Next Monday
```

### Pipeline Operations
```bash
calcdate --expr "today | +1M | endOfMonth"    # Last day of next month
calcdate --expr "now | +2h | round hour"      # Round to nearest hour
```

### Boundary Operations
```bash
calcdate --expr "today | startOfMonth"
calcdate --expr "today | endOfWeek"
calcdate --expr "now | startOfHour"
```

### Using Stdin
```bash
echo "today" | calcdate
echo "today +1w" | calcdate --format=iso
echo "today...+30d" | calcdate --each=1w
```

### Business Day Calculations
```bash
calcdate --expr "today +5bd" --skip-weekends
```

## Format String Translation

### v1.x Format Placeholders
- `%YYYY` - 4-digit year
- `%MM` - 2-digit month
- `%DD` - 2-digit day
- `%hh` - 2-digit hour
- `%mm` - 2-digit minute
- `%ss` - 2-digit second
- `//` - current date
- `::` - current time

### v2.0 Format Options

Use standard Unix date format codes:
- `%Y` - 4-digit year
- `%m` - 2-digit month
- `%d` - 2-digit day
- `%H` - 2-digit hour (24-hour)
- `%M` - 2-digit minute
- `%S` - 2-digit second
- `%Z` - Timezone

Or use predefined formats:
- `iso` - ISO 8601 format
- `sql` - SQL datetime format (default)
- `ts` - Unix timestamp
- `human` - Human-readable format
- `compact` - Compact format

## Common Migration Patterns

### Scripts Using calcdate

#### Before (v1.x):
```bash
#!/bin/bash
START=$(calcdate -b "// ::" -ofmt "%YYYY-%MM-%DD")
END=$(calcdate -b "// ::" -e "+30" -ofmt "%YYYY-%MM-%DD")
echo "SELECT * FROM logs WHERE date BETWEEN '$START' AND '$END'"
```

#### After (v2.0):
```bash
#!/bin/bash
START=$(calcdate --expr "today" --format="%Y-%m-%d")
END=$(calcdate --expr "today +30d" --format="%Y-%m-%d")
echo "SELECT * FROM logs WHERE date BETWEEN '$START' AND '$END'"
```

## Need Help?

- Check the [README](README.md) for comprehensive examples
- Run `calcdate --help` for command-line options
- Report issues at [GitHub Issues](https://github.com/sgaunet/calcdate/issues)