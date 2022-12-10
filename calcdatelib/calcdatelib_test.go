package calcdatelib

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNewDate(t *testing.T) {
	type testCase struct {
		Name           string
		argDate        string
		argIfmt        string
		argTz          string
		expectedString string
		expectedErr    bool
	}

	minNowMinus1 := time.Now().Minute() - 1
	if minNowMinus1 == -1 {
		minNowMinus1 = 59
	}
	hourNowMinus1 := time.Now().Hour() - 1
	if hourNowMinus1 == -1 {
		hourNowMinus1 = 23
	}
	testCases := []testCase{
		{
			Name:           "min=59",
			argDate:        fmt.Sprintf("2022/05/11 02:%d:50", -time.Now().Minute()-1),
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "UTC",
			expectedString: "2022/05/11 01:59:50",
			expectedErr:    false,
		},
		{
			Name:           "wrong tz",
			argDate:        "1982/05/12 12:00:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "%YYYY/%MM/%DD %hh:%mm:%ss",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "ok",
			argDate:        "1982/05/12 12:00:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "1982/05/12 12:00:01",
			expectedErr:    false,
		},
		{
			Name:           "ifmt %YYYY-%MM-%DD %hh:%mm:%ss",
			argDate:        "1982-05-12 12:00:01",
			argIfmt:        "%YYYY-%MM-%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "1982/05/12 12:00:01",
			expectedErr:    false,
		},
		{
			Name:           "ifmt %hh:%mm:%ss",
			argDate:        "12:00:01 120ert12ert29",
			argIfmt:        "%hh:%mm:%ss %YYYYert%MMert%DD",
			argTz:          "",
			expectedString: "0120/12/29 12:00:01",
			expectedErr:    false,
		},
		{
			Name:           "ok",
			argDate:        "1982/05/12 12:-1:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: fmt.Sprintf("1982/05/12 12:%02d:01", minNowMinus1),
			expectedErr:    false,
		},
		{
			Name:           "ok",
			argDate:        "1982/05/12 -1:01:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: fmt.Sprintf("1982/05/12 %02d:01:01", hourNowMinus1),
			expectedErr:    false,
		},
		{
			Name:           "year ko",
			argDate:        "year/05/12 -1:01:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "month ko",
			argDate:        "2020/month/12 -1:01:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "day ko",
			argDate:        "2020/05/day -1:01:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "hour ko",
			argDate:        "2020/05/12 hour:01:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "min ko",
			argDate:        "2020/05/12 -1:min:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "sec ko",
			argDate:        "2020/05/12 -1:01:sec",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "tz ko",
			argDate:        "2020/05/12 -1:01:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "tz",
			expectedString: "",
			expectedErr:    true,
		},
		{
			Name:           "tz ok",
			argDate:        "2020/05/12 15:01:01",
			argIfmt:        "%YYYY/%MM/%DD %hh:%mm:%ss",
			argTz:          "Europe/Paris",
			expectedString: "2020/05/12 15:01:01",
			expectedErr:    false,
		},
		{
			Name:           "different format",
			argDate:        "2020-05-12 15:01:01",
			argIfmt:        "%YYYY-%MM-%DD %hh:%mm:%ss",
			argTz:          "Europe/Paris",
			expectedString: "2020/05/12 15:01:01",
			expectedErr:    false,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			d, err := NewDate(test.argDate, test.argIfmt, test.argTz)
			isError := err != nil
			if (isError) != test.expectedErr {
				t.Errorf("case %s in error", test.Name)
				t.Errorf("expectedString=%s VS result=%s", test.expectedString, d.String())
			}
			if d != nil {
				if d.String() != test.expectedString {
					t.Errorf("case %s in error", test.Name)
					t.Errorf("expectedString=%s VS result=%s", test.expectedString, d.String())
				}
			}
		})
	}
}

func TestTime(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/London")
	ex := time.Date(2003, time.Month(3), 02, 10, 20, 30, 0, loc)
	d, _ := NewDate("2003/3/2 10:20:30", "%YYYY/%MM/%DD %hh:%mm:%ss", "Europe/London")
	if ex.Unix() != d.Time().Unix() {
		t.Errorf("Time should be same")
	}
}

func TestBefore(t *testing.T) {
	d, _ := NewDate("2003/3/2 10:20:30", "%YYYY/%MM/%DD %hh:%mm:%ss", "Europe/London")
	after, _ := NewDate("2003/3/2 10:20:31", "%YYYY/%MM/%DD %hh:%mm:%ss", "Europe/London")
	if !d.Before(after) {
		t.Error("d < after")
	}
	if d.Before(d) {
		t.Error("same date")
	}
}

func TestAdd(t *testing.T) {
	d, _ := NewDate("2003/3/2 23:59:59", "%YYYY/%MM/%DD %hh:%mm:%ss", "Europe/London")
	res, _ := NewDate("2003/3/3 00:00:00", "%YYYY/%MM/%DD %hh:%mm:%ss", "Europe/London")
	d.Add(time.Second)
	if d.Before(res) {
		t.Error("same date")
	}
	if res.Before(d) {
		t.Error("same date")
	}
}

func TestSetBeginDate(t *testing.T) {
	type testCase struct {
		argDate        string
		expectedResult string
	}

	tests := []testCase{
		{
			argDate:        "2003/3/2 23:59:45",
			expectedResult: "2003/03/02 23:59:45",
		},
		{
			argDate:        "2003/3/2 23:59:",
			expectedResult: "2003/03/02 23:59:00",
		},
		{
			argDate:        "2003/3/2 23::",
			expectedResult: "2003/03/02 23:00:00",
		},
		{
			argDate:        "2003/3/2 ::",
			expectedResult: "2003/03/02 00:00:00",
		},
		{
			argDate:        "2003/3/ ::",
			expectedResult: "2003/03/01 00:00:00",
		},
		{
			argDate:        "2003// ::",
			expectedResult: "2003/01/01 00:00:00",
		},
		{
			argDate:        "// ::",
			expectedResult: fmt.Sprintf("%d/01/01 00:00:00", time.Now().Year()),
		},
	}

	for test := range tests {
		d, err := NewDate(tests[test].argDate, "%YYYY/%MM/%DD %hh:%mm:%ss", "Europe/London")
		if err != nil {
			t.Errorf(err.Error())
		}
		d.SetBeginDate()
		if d.String() != tests[test].expectedResult {
			t.Errorf("wrong result=%s  expected=%s", d.String(), tests[test].expectedResult)
		}
	}
}

func TestSetEndDate(t *testing.T) {
	type testCase struct {
		argDate        string
		expectedResult string
	}

	tests := []testCase{
		{
			argDate:        "2003/3/2 21:50:45",
			expectedResult: "2003/03/02 21:50:45",
		},
		{
			argDate:        "2003/3/2 21:50:",
			expectedResult: "2003/03/02 21:50:59",
		},
		{
			argDate:        "2003/3/2 21::",
			expectedResult: "2003/03/02 21:59:59",
		},
		{
			argDate:        "2003/3/2 ::",
			expectedResult: "2003/03/02 23:59:59",
		},
		{
			argDate:        "2003/3/ ::",
			expectedResult: "2003/03/31 23:59:59",
		},
		{
			argDate:        "2003// ::",
			expectedResult: "2003/12/31 23:59:59",
		},
		{
			argDate:        "// ::",
			expectedResult: fmt.Sprintf("%d/12/31 23:59:59", time.Now().Year()),
		},
	}

	for test := range tests {
		d, _ := NewDate(tests[test].argDate, "%YYYY/%MM/%DD %hh:%mm:%ss", "Europe/London")
		d.SetEndDate()
		if d.String() != tests[test].expectedResult {
			t.Errorf("wrong result=%s  expected=%s", d.String(), tests[test].expectedResult)
		}
	}
}

func TestFormat(t *testing.T) {
	type testCase struct {
		Name           string
		argFormat      string
		expectedString string
	}

	argDate := "1982/05/12 12:00:01"
	argIfmt := "%YYYY/%MM/%DD %hh:%mm:%ss"
	argTz := "UTC"
	d, _ := NewDate(argDate, argIfmt, argTz)
	testCases := []testCase{
		{
			Name:           "like ifmt",
			argFormat:      "%YYYY/%MM/%DD %hh:%mm:%ss",
			expectedString: "1982/05/12 12:00:01",
		},
		{
			Name:           "just year",
			argFormat:      "%YYYY",
			expectedString: "1982",
		},
		{
			Name:           "wrong format",
			argFormat:      "%YYY/%M/%D %h:%m:%s",
			expectedString: "%YYY/%M/%D %h:%m:%s",
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			if d.Format(test.argFormat) != test.expectedString {
				t.Errorf("case %s in error: result=%s expected=%s", test.Name, d.Format(test.argFormat), test.expectedString)
			}
		})
	}
}

func TestMain(m *testing.M) {
	//config.TestDatabaseInit()
	fmt.Println("test main")
	ret := m.Run()
	//config.TestDatabaseDestroy()
	os.Exit(ret)
}

func TestGetInterval(t *testing.T) {
	ifmt := "%YYYY/%MM/%DD %hh:%mm:%ss"
	d1, _ := NewDate("// :-5:", ifmt, "")
	d2, _ := NewDate("// ::", ifmt, "")
	d3, _ := NewDate("// :-1:", ifmt, "")
	type args struct {
		d1 *Date
		d2 *Date
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "5 min",
			args: args{
				d1: d1,
				d2: d2,
			},
			want: 5 * time.Minute,
		},
		{
			name: "4 min",
			args: args{
				d1: d1,
				d2: d3,
			},
			want: 4 * time.Minute,
		},
		{
			name: "1 min",
			args: args{
				d1: d3,
				d2: d2,
			},
			want: 1 * time.Minute,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInterval(tt.args.d1, tt.args.d2); got != tt.want {
				t.Errorf("GetInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDate_SetEndMonth(t *testing.T) {
	tests := []struct {
		name string
		date string
		want time.Time
	}{
		{
			name: "end of february",
			date: "2022/02/05 01:02:03",
			want: time.Date(2022, 2, 28, 23, 59, 59, 0, time.UTC),
		},
		{
			name: "end of february",
			date: "2022/02/05 ::",
			want: time.Date(2022, 2, 28, 23, 59, 59, 0, time.UTC),
		},
		{
			name: "end of december",
			date: "2022/12/05 01:02:03",
			want: time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		{
			name: "end of january",
			date: "2022/01/05 01:02:03",
			want: time.Date(2022, 1, 31, 23, 59, 59, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := NewDate(tt.date, "%YYYY/%MM/%DD %hh:%mm:%ss", "UTC")
			result := d.SetEndMonth().Time()
			if result != tt.want {
				t.Errorf("Date.SetEndMonth() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestDate_SetBeginMonth(t *testing.T) {
	tests := []struct {
		name string
		date string
		want time.Time
	}{
		{
			name: "begin of february",
			date: "2022/02/05 01:02:03",
			want: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "begin of december",
			date: "2022/12/05 01:02:03",
			want: time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "begin of january",
			date: "2022/01/05 01:02:03",
			want: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "begin of january",
			date: "2022/01/05 ::",
			want: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := NewDate(tt.date, "%YYYY/%MM/%DD %hh:%mm:%ss", "UTC")
			result := d.SetBeginMonth().Time()
			if result != tt.want {
				t.Errorf("Date.SetBeginMonth() = %v, want %v", result, tt.want)
			}
		})
	}
}
