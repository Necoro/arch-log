package entries

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var ErrNotFound = errors.New("package could not be found remotely")

var (
	timeColor    = color.New(color.FgYellow, color.Bold).Sprint
	summaryColor = color.New(color.Bold).Sprint
	tagColor     = color.New(color.FgGreen).Sprint
)

type Entry struct {
	CommitTime time.Time
	Summary    string
	Message    string
	Author     string
	Tag        string
}

func (e Entry) timeStr() string {
	if e.CommitTime.IsZero() {
		return "(unknown commit time)"
	}
	return e.CommitTime.Local().Format(time.DateTime)
}

func (e Entry) Format() string {
	tag := e.Tag
	if tag != "" {
		tag = tagColor("(" + tag + ")")
	}
	str := fmt.Sprintf("%-24s %s %s", timeColor(e.timeStr()), tag, summaryColor(e.Summary))
	msg := strings.TrimSpace(e.Message)

	if msg != "" {
		str = str + "\n" + msg
	}

	return str
}
