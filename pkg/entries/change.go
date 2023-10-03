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

type Change struct {
	CommitTime time.Time
	Summary    string
	Message    string
	Author     string
	Tag        string
	RepoInfo   string
}

func (c Change) formatTime(format string) string {
	if c.CommitTime.IsZero() {
		return ""
	}
	return c.CommitTime.Local().Format(format)
}

func (c Change) timeStr() string {
	return c.formatTime(time.DateTime)
}

func (c Change) dateStr() string {
	return c.formatTime(time.DateOnly)
}

func (c Change) tagStr() string {
	if c.Tag != "" {
		return "(" + c.Tag + ")"
	}
	return ""
}

func (c Change) repoStr() string {
	if c.RepoInfo != "" {
		return "[" + c.RepoInfo + "]"
	}
	return ""
}

func (c Change) Format() string {
	dateTime := timeColor.Sprintf("%-19s", c.timeStr())

	tag := c.tagStr()
	if tag != "" {
		tag = " " + tagColor.Sprint(tag)
	}

	repo := c.repoStr()
	if repo != "" {
		repo = " " + repoColor.Sprint(repo)
	}

	summary := summaryColor.Sprint(c.Summary)
	str := fmt.Sprintf("%s%s%s %s", dateTime, tag, repo, summary)

	msg := strings.TrimSpace(c.Message)

	if msg != "" {
		str = str + "\n" + msg
	}

	return str
}

func (c Change) ShortFormat(tagLength, repoLength int) string {
	start := startColor.Sprint("*")
	date := timeColor.Sprintf("%10s", c.dateStr())

	tag := ""
	if tagLength > 0 {
		tagLength = tagLength + 2 // parens
		tag = tagColor.Sprintf(" %*s", tagLength, c.tagStr())
	}

	repoInfo := ""
	if repoLength > 0 {
		repoLength = repoLength + 2 // []
		repoInfo = repoColor.Sprintf(" %-*s", repoLength, c.repoStr())
	}

	summary := summaryColor.Sprint(c.Summary)

	msg := ""
	if strings.TrimSpace(c.Message) != "" {
		msg = " [...]"
	}

	return fmt.Sprintf("%s %s%s%s %s%s", start, date, tag, repoInfo, summary, msg)
}
