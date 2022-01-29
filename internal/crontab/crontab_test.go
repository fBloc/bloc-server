package crontab

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCrontabStringValidCheck(t *testing.T) {
	crontabstrMapvalid := map[string]bool{
		"* * * * *":            true,
		"*/1 * * * *":          true,
		"47 08-19 * * *":       true,
		"02 08-19/2 * * *":     true,
		"02 08,12,14-16 * * *": true,
		"* * * * * *":          false,
		"69 * * * *":           false,
		"* 34 * * *":           false,
		"* 14-26 * * *":        false,
	}

	Convey("crontab string valid check", t, func() {
		for str, expect := range crontabstrMapvalid {
			actual := IsCrontabStringValid(str)
			So(actual, ShouldEqual, expect)
		}
	})
}

func TestCrontabBuild(t *testing.T) {
	validCrontabStrs := []string{
		"* * * * *",
		"*/1 * * * *",
		"47 08-19 * * *",
		"02 08-19/2 * * *",
		"02 08,12,14-16 * * *",
	}

	Convey("crontab build check", t, func() {
		diffCr := BuildCrontab("*/5 * * * *")

		for _, str := range validCrontabStrs {
			cr := BuildCrontab(str)
			So(cr, ShouldNotBeNil)
			So(cr.IsZero(), ShouldBeFalse)
			So(cr.IsValid(), ShouldBeTrue)
			So(cr.String(), ShouldNotEqual, "")
			So(cr.String(), ShouldEqual, str)
			So(cr.Equal(diffCr), ShouldBeFalse)
			crB := BuildCrontab(str)
			So(cr.Equal(crB), ShouldBeTrue)
		}
	})
}

func TestCrontabRunMatch(t *testing.T) {
	crontabstrMapNowShouldRun := map[string]bool{
		"* * * * *":   true,
		"*/1 * * * *": true,
	}
	now := time.Now()
	notMatchMinute := now.Minute() - 1
	if notMatchMinute < 0 {
		notMatchMinute = 20
	}
	notMatchHour := now.Hour() - 1
	if notMatchHour < 0 {
		notMatchHour = 20
	}
	notMatchDay := now.Day() - 1
	if notMatchDay < 0 {
		notMatchDay = 20
	}
	crontabstrMapNowShouldRun[fmt.Sprintf("%d * * * *", notMatchMinute)] = false
	crontabstrMapNowShouldRun[fmt.Sprintf("* %d * * *", notMatchHour)] = false
	crontabstrMapNowShouldRun[fmt.Sprintf("* * %d * *", notMatchDay)] = false

	Convey("crontab run now check", t, func() {
		for crontabStr, expect := range crontabstrMapNowShouldRun {
			cr := BuildCrontab(crontabStr)
			if cr != nil {
				actual := cr.TimeMatched(now)
				So(actual, ShouldEqual, expect)
			}
		}
	})
}
