package entries

import (
	"errors"
	"fmt"
	"strings"
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
	str := fmt.Sprintf("%-24s %s", e.timeStr(), e.Summary)
	msg := strings.TrimSpace(e.Message)

	if msg != "" {
		str = str + "\n" + msg
	}

	return str
}
