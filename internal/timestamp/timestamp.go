package timestamp

import (
	"fmt"
	"strconv"
	"time"
)

type Timestamp time.Time

func NewTimeStampFromTime(t time.Time) *Timestamp {
	if t.IsZero() {
		return nil
	}
	tmp := Timestamp(t)
	return &tmp
}

func (t *Timestamp) ToTime() time.Time {
	return time.Time(*t)
}

func (t *Timestamp) IsZero() bool {
	if t == nil {
		return true
	}
	return time.Time(*t).IsZero()
}

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}
	*t = Timestamp(time.Unix(int64(ts), 0))
	return nil
}
