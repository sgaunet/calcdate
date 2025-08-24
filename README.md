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

**The -expr (or -x) parameter is required.**

## Examples

```bash
# Basic date operations
$ calcdate --expr "today"
2024/01/15 00:00:00

$ calcdate --expr "tomorrow"
2024/01/16 00:00:00

$ calcdate --expr "today +1w"
2024/01/22 00:00:00

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
