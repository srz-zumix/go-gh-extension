package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// periodPattern matches a positive integer followed by a single unit letter.
var periodPattern = regexp.MustCompile(`^([0-9]+)([wdhmy])$`)

// ParseTime parses s as a point in time. It accepts RFC3339 timestamps
// (2006-01-02T15:04:05Z07:00), plain dates (2006-01-02, interpreted as that day's start in
// local time), and periods measured back from now (e.g. "2w", "7d", "24h", "30m").
func ParseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}
	if d, err := ParseDuration(s); err == nil {
		return time.Now().Add(-d), nil
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s", s)
}

// ParseDuration parses s as a positive duration expressed as an integer followed by a single
// unit letter: "w" (week), "d" (day), "h" (hour), or "m" (minute) — e.g. "2w", "7d", "24h",
// "30m".
func ParseDuration(s string) (time.Duration, error) {
	n, unitName, err := parsePeriod(s, "wdhm")
	if err != nil {
		return 0, err
	}
	var durationUnit time.Duration
	switch unitName {
	case "w":
		durationUnit = 7 * 24 * time.Hour
	case "d":
		durationUnit = 24 * time.Hour
	case "h":
		durationUnit = time.Hour
	case "m":
		durationUnit = time.Minute
	}
	return time.Duration(n) * durationUnit, nil
}

// fiscalYearStartMonth is the first month of a fiscal year (April, matching
// the Japanese fiscal year convention).
const fiscalYearStartMonth = time.April

// fiscalPeriodPattern matches fiscal year period expressions such as "FY26",
// "FY2026", "FY26H1" and "FY26Q3".
var fiscalPeriodPattern = regexp.MustCompile(`(?i)^FY(\d{2}|\d{4})(H[12]|Q[1-4])?$`)

// fiscalPeriodOffsets maps an H/Q suffix to the number of months after the
// fiscal year start and the span of the period in months.
var fiscalPeriodOffsets = map[string][2]int{
	"":   {0, 12},
	"H1": {0, 6},
	"H2": {6, 6},
	"Q1": {0, 3},
	"Q2": {3, 3},
	"Q3": {6, 3},
	"Q4": {9, 3},
}

// ParsePeriod parses a relative period such as "7d", "3w", "6m" or "1y" into a
// duration. Supported units are d(day), w(week),
// m(month, treated as 30 days) and y(year, treated as 365 days).
func ParsePeriod(period string) (time.Duration, error) {
	n, unitName, err := parsePeriod(period, "dwmy")
	if err != nil {
		return 0, fmt.Errorf("invalid period %q: expected format <N>d|w|m|y", period)
	}

	var duration time.Duration
	switch unitName {
	case "d":
		duration = 24 * time.Hour
	case "w":
		duration = 7 * 24 * time.Hour
	case "m":
		duration = 30 * 24 * time.Hour
	case "y":
		duration = 365 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid period %q: expected format <N>d|w|m|y", period)
	}

	return time.Duration(n) * duration, nil
}

// parsePeriod parses a positive integer and validates its unit against allowedUnits.
func parsePeriod(period, allowedUnits string) (int, string, error) {
	m := periodPattern.FindStringSubmatch(period)
	if m == nil || !strings.ContainsRune(allowedUnits, rune(m[2][0])) {
		return 0, "", fmt.Errorf("invalid period format: %s", period)
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0, "", fmt.Errorf("invalid period: %s", period)
	}
	return n, m[2], nil
}

// IsFiscalPeriod reports whether period uses the fiscal-year period format.
func IsFiscalPeriod(period string) bool {
	return fiscalPeriodPattern.MatchString(period)
}

// ParseFiscalPeriod parses a fiscal year period such as "FY26" (the full
// fiscal year), "FY26H1"/"FY26H2" (half), or "FY26Q1".."FY26Q4" (quarter)
// into a (since, until) window. The fiscal year starts in April.
func ParseFiscalPeriod(period string) (time.Time, time.Time, error) {
	m := fiscalPeriodPattern.FindStringSubmatch(period)
	if m == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid fiscal period %q: expected format FY<YY>[H1|H2|Q1..Q4]", period)
	}

	year, err := strconv.Atoi(m[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid fiscal period %q: %w", period, err)
	}
	if len(m[1]) == 2 {
		year += 2000
	}

	offsets := fiscalPeriodOffsets[strings.ToUpper(m[2])]
	fyStart := time.Date(year, fiscalYearStartMonth, 1, 0, 0, 0, 0, time.Local)
	since := fyStart.AddDate(0, offsets[0], 0)
	until := since.AddDate(0, offsets[1], 0).Add(-time.Nanosecond)
	return since, until, nil
}
