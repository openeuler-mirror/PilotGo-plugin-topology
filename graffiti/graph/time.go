package graph

import (
	"encoding/json"
	"strconv"
	"time"
)

// Time describes time type used in the graph
type Time time.Time

// UnixMilli returns the time in milliseconds since January 1, 1970
func (t Time) UnixMilli() int64 {
	return time.Time(t).UnixNano() / int64(time.Millisecond)
}

// MarshalJSON custom marshalling function
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.UnixMilli())
}

// UnmarshalJSON custom unmarshalling function
func (t *Time) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	ms, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	*t = Time(time.Unix(0, ms*int64(time.Millisecond)))

	return nil
}

// IsZero returns is empty or not
func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}

// TimeNow creates a Time with now local time
func TimeNow() Time {
	return Time(time.Now())
}

// TimeUTC creates a Time with now UTC
func TimeUTC() Time {
	return Time(time.Now().UTC())
}

// Unix returns Time for given sec, nsec
func Unix(sec int64, nsec int64) Time {
	return Time(time.Unix(sec, nsec))
}
