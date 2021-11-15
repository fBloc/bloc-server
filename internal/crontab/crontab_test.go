package crontab

import (
	"fmt"
	"testing"
	"time"
)

func TestCrontab(t *testing.T) {
	crontabstrMapvalid := make(map[string]bool, 10)
	crontabstrMapvalid["* * * * *"] = true
	crontabstrMapvalid["*/1 * * * *"] = true
	crontabstrMapvalid["47 08-19 * * *"] = true
	crontabstrMapvalid["02 08-19/2 * * *"] = true
	crontabstrMapvalid["* * * * * *"] = false
	crontabstrMapvalid["69 * * * *"] = false
	crontabstrMapvalid["* 34 * * *"] = false

	now := time.Now()
	for k, v := range crontabstrMapvalid {
		cr := BuildCrontab(k)
		if v == true && cr.IsZero() {
			t.Errorf("%v should be valid", k)
		}
		if v == false && cr != nil {
			t.Errorf("%v should not be valid", k)
		}
		if cr != nil {
			run := cr.TimeMatched(now)
			if run {
				fmt.Printf("\tRun Now! %v\n", k)
			}
		}
	}
}
