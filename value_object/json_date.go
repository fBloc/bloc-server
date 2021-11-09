package value_object

import (
	"bytes"
	"fmt"
	"time"
)

var cstSh, _ = time.LoadLocation("Asia/Shanghai")

const timeFormat = "2006-01-02 15:04:05"

type JsonDate struct {
	Time time.Time
}

func NewJsonDateFromStr(t string) (JsonDate, error) {
	timeIns, err := time.Parse(timeFormat, t)
	if err != nil {
		return JsonDate{}, nil
	}
	return NewJsonDate(timeIns), nil
}

func NewJsonDate(t time.Time) JsonDate {
	return JsonDate{Time: t}
}

func NowJsonDate() JsonDate {
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
	return t.Time.Format(timeFormat)
}

func (t *JsonDate) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, "\"")
	ext, err := time.ParseInLocation(timeFormat, string(b), cstSh)
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
