[![GitHub release](https://img.shields.io/github/release/sgaunet/calcdate.svg)](https://github.com/sgaunet/calcdate/releases/latest)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/calcdate/total)
[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/calcdate)](https://goreportcard.com/report/github.com/sgaunet/calcdate)
![Test Coverage](https://raw.githubusercontent.com/wiki/sgaunet/calcdate/coverage-badge.svg)
[![linter](https://github.com/sgaunet/calcdate/actions/workflows/coverage.yml/badge.svg)](https://github.com/sgaunet/calcdate/actions/workflows/coverage.yml)
[![coverage](https://github.com/sgaunet/calcdate/actions/workflows/coverage.yml/badge.svg)](https://github.com/sgaunet/calcdate/actions/workflows/coverage.yml)
[![Snapshot Build](https://github.com/sgaunet/calcdate/actions/workflows/snapshot.yml/badge.svg)](https://github.com/sgaunet/calcdate/actions/workflows/snapshot.yml)
[![Release Build](https://github.com/sgaunet/calcdate/actions/workflows/release.yml/badge.svg)](https://github.com/sgaunet/calcdate/actions/workflows/release.yml)
[![GoDoc](https://godoc.org/github.com/sgaunet/calcdate?status.svg)](https://godoc.org/github.com/sgaunet/calcdate)
[![License](https://img.shields.io/github/license/sgaunet/calcdate.svg)](LICENSE)

# calcdate

calcdate is a utility to make some basic operation on date. It's useful when need to calculate a range of date in order to make database request.

## Expression Syntax

calcdate uses a simple, intuitive expression syntax:

```bash
# Simple date calculations
calcdate --expr "today +1d"                    # Tomorrow
calcdate --expr "now +2h"                      # 2 hours from now
calcdate --expr "yesterday"                    # Yesterday

# Pipeline operations
calcdate --expr "today | +1M | endOfMonth"     # Last day of next month
calcdate --expr "now | +2h | round hour"       # 2 hours from now, rounded

# Range operations with iterations
calcdate --expr "today...+7d" --each=1d        # Each day for next week
calcdate --expr "today...+30d" --each=1w --transform='$begin +8h, $end +20h'  # Business hours each week

# Different output formats
calcdate --expr "tomorrow" --format=iso        # 2024-01-16T00:00:00Z
calcdate --expr "today" --format=sql           # 2024-01-15 00:00:00
calcdate --expr "now" --format=ts              # 1705331400
```

### Quick Reference

| Expression | Result |
|------------|--------|
| `today` | Start of today (00:00:00) |
| `now` | Current date and time |
| `tomorrow` | Start of tomorrow |
| `today +1w` | One week from today |
| `today \| endOfMonth` | Last day of current month |
| `today...+7d` | Range from today to 7 days from now |

## Usage

```
Usage of calcdate:
  -each string
        Iteration interval for ranges (e.g., '1d', '1w', '1M')
  -expr string
        Date expression (e.g., 'today +1d', 'now | +2h | round hour', 'today...+7d')
  -f string
        Output format (short form)
  -format string
        Output format: iso, sql, ts, human, compact, or Unix date format
        (e.g., '%Y-%m-%d %H:%M:%S', '%Y-%m-%d %H:%M:%S %Z')
  -list-tz
        List timezones
  -skip-weekends
        Skip weekend days in iterations
  -t string
        Transform expression (short form)
  -transform string
        Transform expression for iterations (e.g., '$begin +8h, $end +20h')
  -tz string
        Input timezone (default "Local")
  -v    Get version
  -x string
        Date expression (short form)
```

**The -expr (or -x) parameter is required, or provide the expression via stdin.**

## Examples

```bash
# Basic date operations
$ calcdate --expr "today"
2024/01/15 00:00:00

$ calcdate --expr "tomorrow"
2024/01/16 00:00:00

$ calcdate --expr "today +1w"
2024/01/22 00:00:00

# Using stdin (pipe expressions)
$ echo "today" | calcdate
2024-01-15 00:00:00

$ echo "tomorrow" | calcdate --format iso
2024-01-16T00:00:00Z

$ echo "today +1w" | calcdate --format sql
2024-01-22 00:00:00

# Boundary operations
$ calcdate --expr "today | endOfMonth"
2024/01/31 23:59:59

$ calcdate --expr "today | startOfWeek"
2024/01/15 00:00:00

# Date ranges with iterations
$ calcdate --expr "today...+7d" --each=1d
2024/01/15 00:00:00 - 2024/01/16 00:00:00
2024/01/16 00:00:00 - 2024/01/17 00:00:00
...

# Business hours (8am to 8pm each day)
$ calcdate --expr "today...+7d" --each=1d --transform='$begin +8h, $end +20h'
2024/01/15 08:00:00 - 2024/01/15 20:00:00
2024/01/16 08:00:00 - 2024/01/16 20:00:00
...

# Different output formats
$ calcdate --expr "today" --format=sql
2024-01-15 00:00:00

$ calcdate --expr "now" --format=ts
1705331400

$ calcdate --expr "tomorrow" --format=iso
2024-01-16T00:00:00Z

# Stdin with ranges and iterations
$ echo "today...+7d" | calcdate --each=1d --format=sql | head -3
2024-01-15 00:00:00 - 2024-01-16 00:00:00
2024-01-16 00:00:00 - 2024-01-17 00:00:00
2024-01-17 00:00:00 - 2024-01-18 00:00:00

# Advanced stdin usage examples

## Timezone conversions
# Convert current UTC time to CET timezone
$ date -u "+%Y-%m-%dT%H:%M:%SZ" | calcdate --tz CET --format="%Y-%m-%d %H:%M:%S %Z"
2024-01-15 15:30:45 CET

# Parse date in EST and output in UTC
$ echo "2024-01-15 10:30:00 EST" | calcdate --format iso
2024-01-15T15:30:00Z

## Scripting and automation
# Get business hours for next week (Monday-Friday, 9am-5pm)
$ echo "today | startOfWeek +1w...endOfWeek +1w" | calcdate --each=1d --skip-weekends --transform='$begin +9h, $end +17h'

# Generate database-friendly timestamps for the last 30 days
$ echo "today -30d...today" | calcdate --each=1d --format=ts

# Calculate deployment windows (every Sunday at 2 AM for next 3 months)
$ echo "today | startOfWeek +7d...+3M" | calcdate --each=1w --transform='$begin +2h' --format=iso

## Pipeline chaining with other tools
# Generate log rotation dates and create directories
$ echo "today...+365d" | calcdate --each=1M --format=compact | xargs -I {} mkdir -p logs/{}

# Create backup schedule for weekdays only
$ echo "today...+30d" | calcdate --each=1d --skip-weekends --format='backup-%Y%m%d'

# Generate monitoring time ranges
$ echo "today -7d | startOfDay...today | endOfDay" | calcdate --each=1h --format='%Y-%m-%d %H:00:00'

## Working with different date formats
# Parse ISO date and add business days
$ echo "2024-12-20T09:00:00Z +5bd" | calcdate --skip-weekends --format=human

# Convert Unix timestamp to readable format in specific timezone
$ echo "@1705331400 +1d" | calcdate --tz America/New_York --format="%A, %B %d, %Y at %I:%M %p"

# Parse various input formats
$ echo "Dec 25, 2024" | calcdate --format=iso
$ echo "2024-12-25" | calcdate --format=human
$ echo "next Monday" | calcdate --format=compact
```


# Install

## Option 1

* Download the release
* Install the binary in /usr/local/bin 

## Option 2: With brew

```
brew tap sgaunet/homebrew-tools
brew install sgaunet/tools/calcdate
```
