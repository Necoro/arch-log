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
	timeColor    = color.New(color.FgYellow, color.Bold)
	summaryColor = color.New(color.Bold)
	tagColor     = color.New(color.FgGreen)
	repoColor    = color.New(color.FgYellow)
	startColor   = color.New(color.FgGreen, color.Bold)
)

type Entry struct {
	CommitTime time.Time
	Summary    string
	Message    string
	Author     string
	Tag        string
	RepoInfo   string
}

func (e Entry) formatTime(format string) string {
	if e.CommitTime.IsZero() {
		return ""
	}
	return e.CommitTime.Local().Format(format)
}

func (e Entry) timeStr() string {
	return e.formatTime(time.DateTime)
}

func (e Entry) dateStr() string {
	return e.formatTime(time.DateOnly)
}

func (e Entry) tagStr() string {
	if e.Tag != "" {
		return "(" + e.Tag + ")"
	}
	return ""
}

func (e Entry) repoStr() string {
	if e.RepoInfo != "" {
		return "[" + e.RepoInfo + "]"
	}
	return ""
}

func (e Entry) Format() string {
	dateTime := timeColor.Sprintf("%-19s", e.timeStr())

	tag := e.tagStr()
	if tag != "" {
		tag = " " + tagColor.Sprint(tag)
	}

	repo := e.repoStr()
	if repo != "" {
		repo = " " + repoColor.Sprint(repo)
	}

	summary := summaryColor.Sprint(e.Summary)
	str := fmt.Sprintf("%s%s%s %s", dateTime, tag, repo, summary)

	msg := strings.TrimSpace(e.Message)

	if msg != "" {
		str = str + "\n" + msg
	}

	return str
}

func (e Entry) ShortFormat(tagLength, repoLength int) string {
	start := startColor.Sprint("*")
	date := timeColor.Sprintf("%10s", e.dateStr())

	tag := ""
	if tagLength > 0 {
		tagLength = tagLength + 2 // parens
		tag = tagColor.Sprintf(" %*s", tagLength, e.tagStr())
	}

	repoInfo := ""
	if repoLength > 0 {
		repoLength = repoLength + 2 // []
		repoInfo = repoColor.Sprintf(" %-*s", repoLength, e.repoStr())
	}

	summary := summaryColor.Sprint(e.Summary)

	msg := ""
	if strings.TrimSpace(e.Message) != "" {
		msg = " [...]"
	}

	return fmt.Sprintf("%s %s%s%s %s%s", start, date, tag, repoInfo, summary, msg)
}
