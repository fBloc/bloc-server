package json_date

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

var cstSh, _ = time.LoadLocation("Asia/Shanghai")

const timeFormat = time.RFC3339

type JsonDate struct {
	Time time.Time
}

func New(t time.Time) JsonDate {
	return JsonDate{Time: t}
}

func Now() JsonDate {
	return JsonDate{Time: time.Now()}
}

func (t *JsonDate) IsZero() bool {
	if t == nil {
		return true
	}
	return t.Time.IsZero()
}

func (t *JsonDate) String() string {
	if t.IsZero() {
		return ""
	}
	return t.Time.Format("2006-01-02 15:04:05")
}

func (t *JsonDate) UnmarshalJSON(b []byte) error {
	s := string(bytes.Trim(b, "\""))
	ext, err := time.Parse(timeFormat, s)
	if !strings.HasSuffix(s, "Z") {
		utcGap, err := time.Parse("15:04", s[len(s)-5:])
		if err != nil {
			return err
		}

		if strings.Contains(s, "+") {
			ext = ext.Add(-time.Duration(utcGap.Hour()) * time.Hour).Add(-time.Duration(utcGap.Second()) * time.Second)
		} else {
			ext = ext.Add(time.Duration(utcGap.Hour()) * time.Hour).Add(time.Duration(utcGap.Second()) * time.Second)
		}
	}
	if err != nil {
		return err
	}
	*t = JsonDate{ext}
	return nil
}

func (t JsonDate) MarshalJSON() ([]byte, error) {
	var stamp string
	if t.IsZero() {
		stamp = "\"\""
	} else {
		stamp = fmt.Sprintf(
			"\"%s\"",
			time.Time(t.Time).In(cstSh).Format(timeFormat))
	}
	return []byte(stamp), nil
}
