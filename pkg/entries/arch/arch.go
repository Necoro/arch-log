package arch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
)

type entry struct {
	Title     string
	Timestamp string `json:"created_at"`
	Author    string `json:"author_name"`
	Message   string
}

func (e entry) convertTime() time.Time {
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

func (e entry) cleanedMessage() string {
	if e.Message == e.Title {
		return ""
	}

	headerEnd := strings.Index(e.Message, "\n\n")
	if headerEnd == -1 {
		return e.Message
	}
	return e.Message[headerEnd+1:]
}

func fetch(url string) ([]entry, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, entries.ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetching %s: Server returned status %s", url, resp.Status)
	}

	log.Debugf("Fetching from Arch (%s) successful.", url)

	var jsonEntries []entry
	d := json.NewDecoder(resp.Body)
	if err := d.Decode(&jsonEntries); err != nil {
		return nil, err
	}
	return jsonEntries, nil
}

func buildUrl(pkg string) string {
	repoName := url.QueryEscape("archlinux/packaging/packages/" + pkg)

	return "https://gitlab.archlinux.org/api/v4/projects/" + repoName + "/repository/commits"
}

func convert(jsonEntries []entry) []entries.Entry {
	entryList := make([]entries.Entry, len(jsonEntries))
	for i, jsonE := range jsonEntries {
		log.Debugf("Fetched entry %+v", jsonE)

		entryList[i] = entries.Entry{
			CommitTime: jsonE.convertTime(),
			Author:     jsonE.Author,
			Summary:    jsonE.Title,
			Message:    jsonE.cleanedMessage(),
		}
	}
	return entryList
}

//goland:noinspection GoImportUsedAsName
func GetEntries(pkg string) ([]entries.Entry, error) {
	url := buildUrl(pkg)
	jsonEntries, err := fetch(url)
	if err != nil {
		return nil, err
	}
	return convert(jsonEntries), nil
}
