package arch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/http"
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
	result, err := http.Fetch(url)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	log.Debugf("Fetching from Arch (%s) successful.", url)

	var jsonEntries []entry
	d := json.NewDecoder(result)
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
	basePkg, err := determineBasePkg(pkg)
	if err != nil {
		return nil, err
	}
	if basePkg != pkg {
		log.Printf("Mapped pkg '%s' to pkgbase '%s'", pkg, basePkg)
	}

	url := buildUrl(basePkg)
	jsonEntries, err := fetch(url)
	if err != nil {
		if errors.Is(err, entries.ErrNotFound) {
			err = fmt.Errorf("package not found on Gitlab, even though it exists on Packages @ Arch")
		}
		return nil, err
	}
	return convert(jsonEntries), nil
}
