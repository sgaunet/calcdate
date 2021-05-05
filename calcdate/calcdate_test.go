package calcdate

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCheckDateFormat1(t *testing.T) {
	if !CheckDateFormat("-1982/05/12 12:00:01", "YYYY/MM/DD hh:mm:ss") {
		t.Error("Date format is acceptable")
	}
}

func TestCheckDateFormat2(t *testing.T) {
	if !CheckDateFormat("1982/05/12 12:00:01", "YYYY/MM/DD hh:mm:ss") {
		t.Error("Date has a good format")
	}
}

func TestCheckDateFormat3(t *testing.T) {
	if CheckDateFormat("1982/13/12 12:00:01", "YYYY/MM/DD hh:mm:ss") {
		t.Error("Date has a wrong format")
	}
}

func TestCheckDateFormat4(t *testing.T) {
	if CheckDateFormat("1982/12/32 12:00:01", "YYYY/MM/DD hh:mm:ss") {
		t.Error("Date has a wrong format")
	}
}

func TestCheckDateFormat5(t *testing.T) {
	if CheckDateFormat("1982/12/31 102:00:01", "YYYY/MM/DD hh:mm:ss") {
		t.Error("Date has a wrong format")
	}
}

func TestCheckDateFormat6(t *testing.T) {
	if !CheckDateFormat("//-10 ::", "YYYY/MM/DD hh:mm:ss") {
		t.Error("Date has a wrong format")
	}
}

func TestCreeDateVide(t *testing.T) {
	if _, err := CreateDate("", "", "", true, false); err == nil {
		t.Error("Should be wrong, empty date given")
	}
}

func TestCreeDate1(t *testing.T) {
	if _, err := CreateDate("2020/08/22 18:00:", "YYYY/MM/DD hh:mm:ss", "UTC", false, false); err != nil {
		t.Error("Should be ok")
	}
}

func TestCreateDate(t *testing.T) {
	if _, err := CreateDate("2020/13/22 18:00", "YYYY/MM/DD hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("Month doesn't exist")
	}
}

func TestCreeDate3(t *testing.T) {
	if _, err := CreateDate("2020/08/22 18:00", "YYYY/MM/ hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("ifmt is not good")
	}
}

func TestCreeDate4(t *testing.T) {
	if _, err := CreateDate("// ::", "YYYY/MM/SS hh:mm:ss", "UTC", true, false); err != nil {
		t.Error("CreeDate now")
	}
}

func TestCreeDate5(t *testing.T) {
	if _, err := CreateDate("-1/-1/-1 -1:-1:-1", "YYYY/MM/SS hh:mm:ss", "UTC", false, false); err != nil {
		t.Error("CreeDate -1/-1/-1 -1:-1:-1")
	}
}

func TestCreeDate6(t *testing.T) {
	if _, err := CreateDate("a/-1/-1 -1:-1:-1", "YYYY/MM/SS hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("CreeDate a/-1/-1 -1:-1:-1")
	}
}

func TestCreeDate7(t *testing.T) {
	if _, err := CreateDate("-1/m/-1 -1:-1:-1", "YYYY/MM/SS hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("CreeDate -1/m/-1 -1:-1:-1")
	}
}

func TestCreeDate8(t *testing.T) {
	if _, err := CreateDate("-1/-1/j -1:-1:-1", "YYYY/MM/DD hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("CreeDate -1/-1/j -1:-1:-1")
	}
}

func TestCreeDate9(t *testing.T) {
	if _, err := CreateDate("-1/-1/-1 HH:-1:-1", "YYYY/MM/SS hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("CreeDate -1/-1/-1 HH:-1:-1")
	}
}

func TestCreeDate10(t *testing.T) {
	if _, err := CreateDate("-1/-1/-1 -1:MM:-1", "YYYY/MM/SS hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("CreeDate -1/-1/-1 -1:MM:-1")
	}
}
func TestCreeDate11(t *testing.T) {
	if _, err := CreateDate("-1/-1/-1 -1:-1:SS", "YYYY/MM/SS hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("CreeDate -1/-1/-1 -1:-1:SS")
	}
}

func TestCreeDate12(t *testing.T) {
	if _, err := CreateDate("2020/14/22 18:00", "YYYY/MM/ hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("Month is incorrect")
	}
}

func TestCreeDate13(t *testing.T) {
	if _, err := CreateDate("2020/12/32 18:00", "YYYY/MM/ hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("Day is incorrect")
	}
}

func TestCreeDate14(t *testing.T) {
	if _, err := CreateDate("2020/10/22 28:00", "YYYY/MM/ hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("Hour is incorrect")
	}
}

func TestCreeDate15(t *testing.T) {
	if _, err := CreateDate("2020/11/22 18:60", "YYYY/MM/ hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("Month is incorrect")
	}
}

func TestCreeDate16(t *testing.T) {
	if _, err := CreateDate("2020/10/22 18:00:60", "YYYY/MM/ hh:mm:ss", "UTC", false, false); err == nil {
		t.Error("Second is incorrect")
	}
}

func TestCreeDateWrongTZ(t *testing.T) {
	if _, err := CreateDate("// ::", "YYYY/MM/SS hh:mm:ss", "UTC456", false, false); err == nil {
		t.Error("TZ is wrong")
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

func TestApplyFormat1(t *testing.T) {
	var now = time.Now()
	fmtstr := "YYYY/MM/DD hh:mm:ss"
	res := ApplyFormat(fmtstr, now)
	resWanted := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d", now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second())
	if res != resWanted {
		t.Error("Error Applyformat")
	}
}

func TestApplyFormat2(t *testing.T) {
	var now = time.Now()
	fmtstr := "YYYY/DD/MM mm:hh:ss"
	res := ApplyFormat(fmtstr, now)
	resWanted := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d", now.Year(), now.Day(), int(now.Month()), now.Minute(), now.Hour(), now.Second())
	if res != resWanted {
		t.Error("Error Applyformat")
	}
}

func TestApplyFormat3(t *testing.T) {
	tz := "Local"
	ifmt := "YYYY/MM/DD"
	ofmt := "YYYY-MM-DD"

	begindate, err := CreateDate("//", ifmt, tz, true, false)
	if err != nil {
		t.Error("Error CreateDAte")
	}
	begind := ApplyFormat(ofmt, begindate)
	if len(begind) != 10 {
		t.Error(begind)
		t.Error("Error Applyformat")
	}
}

func TestDoubleReplace(t *testing.T) {
	res := DoubleReplace("YYYY/MM/DD", "YYYY", "%04d", 2021)
	if res != "2021/MM/DD" {
		t.Error("Error DoubleReplace")
	}
}

func TestDoubleReplace2(t *testing.T) {
	res := DoubleReplace("YYYY/MM/DD", "HH", "%02d", 10)
	if res != "YYYY/MM/DD" {
		t.Error("Error DoubleReplace")
	}
}

func TestMain(m *testing.M) {
	//config.TestDatabaseInit()
	fmt.Println("test main")
	ret := m.Run()
	//config.TestDatabaseDestroy()
	os.Exit(ret)
}
