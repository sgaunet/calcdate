package calcdatelib

import (
	"fmt"
	"testing"
	"time"
)

func TestDoubleReplace(t *testing.T) {
	res := doubleReplace("%YYYY/%MM/%DD", "%YYYY", "%04d", 2021)
	if res != "2021/%MM/%DD" {
		t.Error("Error doubleReplace")
	}
}

func TestDoubleReplace2(t *testing.T) {
	res := doubleReplace("YYYY/MM/DD", "HH", "%02d", 10)
	if res != "YYYY/MM/DD" {
		t.Error("Error doubleReplace")
	}
}

func TestConvertStdFormatToGolang(t *testing.T) {
	res := convertStdFormatToGolang("%YYYY %MM %DD %hh %mm %ss")
	if res != "2006 01 02 15 04 05" {
		t.Error("Error in function convertStdFormatToGolang")
	}
}

func TestDayInMonth1(t *testing.T) {
	if DayInMonth(2020, 1) != 31 {
		t.Error("01/2020 => 31 days VS", DayInMonth(2020, 1))
	}
}

func TestDayInMonth2(t *testing.T) {
	if DayInMonth(2020, 2) != 29 {
		t.Error("02/2020 => 29 days VS", DayInMonth(2020, 2))
	}
}

func TestDayInMonth3(t *testing.T) {
	if DayInMonth(2020, 12) != 31 {
		t.Error("02/2020 => 31 days VS", DayInMonth(2020, 12))
	}
}

func TestFindNoOfDays(t *testing.T) {
	type testCase struct {
		year           int
		expectedResult bool
	}

	testCases := []testCase{
		{
			year:           2024,
			expectedResult: true,
		},
		{
			year:           2022,
			expectedResult: false,
		},
		{
			year:           2020,
			expectedResult: true,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%d", test.year), func(t *testing.T) {
			res := IsLeapYear(test.year)
			if res != test.expectedResult {
				t.Errorf("case %d in error", test.year)
			}
		})
	}
}

func TestIsLeapYear(t *testing.T) {
	type testCase struct {
		year           int
		expectedResult bool
	}

	testCases := []testCase{
		{
			year:           2024,
			expectedResult: true,
		},
		{
			year:           2022,
			expectedResult: false,
		},
		{
			year:           2020,
			expectedResult: true,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%d", test.year), func(t *testing.T) {
			res := IsLeapYear(test.year)
			if res != test.expectedResult {
				t.Errorf("case %d in error", test.year)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	argDate := "1982/05/12 12:00:01"
	argIfmt := "%YYYY/%MM/%DD %hh:%mm:%ss"
	argTz := ""
	d, _ := NewDate(argDate, argIfmt, argTz)

	tmpl := "{{ .BeginTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - {{ .EndTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} \\o/"
	// tmpl= "{{ .BeginTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - {{ .EndTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} {{ .BeginTime.Unix }} {{ .EndTime.Unix }}"
	out, _ := RenderTemplate(tmpl, d.Time(), d.Time())
	expected := "1982/05/12 12:00:01 - 1982/05/12 12:00:01 \\o/"
	if out != expected {
		t.Error("problem in render")
	}
}

func TestRenderTemplateMinusOneSecond(t *testing.T) {
	argDate := "1982/05/12 12:00:01"
	argIfmt := "%YYYY/%MM/%DD %hh:%mm:%ss"
	argTz := ""
	d, _ := NewDate(argDate, argIfmt, argTz)

	tmpl := "{{ .BeginTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - {{ (MinusOneSecond .EndTime).Format \" %hh:%mm:%ss \" }} \\o/"
	// tmpl= "{{ .BeginTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - {{ .EndTime.Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} {{ .BeginTime.Unix }} {{ .EndTime.Unix }}"
	out, _ := RenderTemplate(tmpl, d.Time(), d.Time())
	expected := "1982/05/12 12:00:01 -  12:00:00  \\o/"
	if out != expected {
		t.Errorf("out=%s  expected=%s", out, expected)
	}
}

func TestRenderTemplateErr(t *testing.T) {
	argDate := "1982/05/12 12:00:01"
	argIfmt := "%YYYY/%MM/%DD %hh:%mm:%ss"
	argTz := ""
	d, _ := NewDate(argDate, argIfmt, argTz)

	tmpl := "{{ .Format \"%YYYY/%MM/%DD %hh:%mm:%ss\" }} - {{ (MinusOneSecond .EndTime).Format \" %hh:%mm:%ss \" }} \\o/"
	_, err := RenderTemplate(tmpl, d.Time(), d.Time())
	if err == nil {
		t.Errorf("template should not work")
	}
}

func TestAddDays(t *testing.T) {
	ti := time.Date(2022, time.Month(2), 2, 0, 0, 0, 0, time.UTC)
	tExpected := time.Date(2022, time.Month(2), 12, 0, 0, 0, 0, time.UTC)
	if !IsSameDay(AddDays(ti, 10), tExpected) {
		t.Error("time should be same")
	}
}

func TestIsSameDay(t *testing.T) {
	ti := time.Date(2022, time.Month(2), 2, 0, 0, 0, 0, time.UTC)
	if IsSameDay(ti, time.Now()) {
		t.Error("time should not be same")
	}
}

func TestStartOfDay(t *testing.T) {
	a := time.Date(2022, time.Month(2), 2, 23, 15, 10, 5, time.UTC)
	expected := time.Date(2022, time.Month(2), 2, 0, 0, 0, 0, time.UTC)
	b := StartOfDay(a, time.UTC)
	if !IsSameDay(b, expected) {
		t.Error("time should be same")
	}
}

func TestEndOfDay(t *testing.T) {
	a := time.Date(2022, time.Month(2), 2, 23, 15, 10, 5, time.UTC)
	expected := time.Date(2022, time.Month(2), 2, 23, 59, 59, 5, time.UTC)
	b := EndOfDay(a, time.UTC)
	if !IsSameDay(b, expected) {
		t.Error("time should be same")
	}
}

func TestDiffInDays(t *testing.T) {
	a := time.Date(2022, time.Month(2), 2, 23, 15, 10, 5, time.UTC)
	b := time.Date(2022, time.Month(2), 12, 22, 59, 59, 5, time.UTC)
	diff := DiffInDays(a, b)
	expected := 9
	if diff != expected {
		t.Errorf("wrong result: diff=%d expected=%d", diff, expected)
	}
}
