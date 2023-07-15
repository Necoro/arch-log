package entries

import (
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("package could not be found remotely")

type Entry struct {
	CommitTime time.Time
	Summary    string
	Message    string
	Author     string
}

func (e Entry) timeStr() string {
	if e.CommitTime.IsZero() {
		return "(unknown commit time)"
	}
	return e.CommitTime.Local().Format(time.DateTime)
}

func (e Entry) Format() string {
	return fmt.Sprintf("%-24s %s", e.timeStr(), e.Summary)
}
