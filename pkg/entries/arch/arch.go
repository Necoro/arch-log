package arch

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/http"
	"github.com/Necoro/arch-log/pkg/log"
)

type commit struct {
	Title     string
	Timestamp string `json:"created_at"`
	Author    string `json:"author_name"`
	Message   string
	Id        string
}

type tag struct {
	Name   string
	Commit struct{ Id string }
}

func (e commit) convertTime() time.Time {
	if e.Timestamp == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
		log.Warnf("Problem parsing time '%s' -- ignoring: %v.", e.Timestamp, err)
		return time.Time{}
	} else {
		return t
	}
}

func (e commit) cleanedMessage() string {
	if e.Message == e.Title {
		return ""
	}

	headerEnd := strings.Index(e.Message, "\n\n")
	if headerEnd == -1 {
		return e.Message
	}
	return e.Message[headerEnd+1:]
}

func fetch(url string, jsonEntries any) error {
	result, err := http.Fetch(url)
	if err != nil {
		return err
	}
	defer result.Close()

	log.Debugf("Fetching from Arch (%s) successful.", url)

	d := json.NewDecoder(result)
	if err := d.Decode(jsonEntries); err != nil {
		return err
	}
	return nil
}

func buildCommitsUrl(pkg string) string {
	return buildUrl(pkg, "commits")
}

func buildTagsUrl(pkg string) string {
	return buildUrl(pkg, "tags")
}

func buildUrl(pkg, action string) string {
	repoName := url.QueryEscape("archlinux/packaging/packages/" + pkg)

	return "https://gitlab.archlinux.org/api/v4/projects/" + repoName + "/repository/" + action
}

func groupTag(tags []tag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[t.Commit.Id] = t.Name
	}

	return m
}

func convert(commits []commit, tags []tag) []entries.Entry {
	entryList := make([]entries.Entry, len(commits))
	tagMap := groupTag(tags)

	for i, c := range commits {
		log.Debugf("Fetched commit %+v", c)

		entryList[i] = entries.Entry{
			CommitTime: c.convertTime(),
			Author:     c.Author,
			Summary:    c.Title,
			Message:    c.cleanedMessage(),
			Tag:        tagMap[c.Id],
		}
	}
	return entryList
}

//goland:noinspection GoImportUsedAsName
func GetEntries(pkg string) ([]entries.Entry, error) {
	basePkg, err := determineBasePkg(pkg)
	if err != nil {
		return nil, err
	}
	if basePkg != pkg {
		log.Printf("Mapped pkg '%s' to pkgbase '%s'", pkg, basePkg)
	}

	url := buildCommitsUrl(basePkg)
	var commits []commit
	if err := fetch(url, &commits); err != nil {
		return nil, err
	}

	url = buildTagsUrl(basePkg)
	var tags []tag
	if err := fetch(url, &tags); err != nil {
		return nil, err
	}

	return convert(commits, tags), nil
}
