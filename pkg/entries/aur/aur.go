package aur

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
)

type feed struct {
	Entries []entry `xml:"entry"`
}
type entry struct {
	Title   string `xml:"title"`
	Updated string `xml:"updated"`
	Author  string `xml:"author>name"`
}

func (e entry) convertTime() time.Time {
	if e.Updated == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, e.Updated); err != nil {
		log.Warnf("Problem parsing time '%s' -- ignoring: %v.", e.Updated, err)
		return time.Time{}
	} else {
		return t
	}
}

func buildUrl(pkg string) string {
	return "https://aur.archlinux.org/cgit/aur.git/atom/?h=" + pkg
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

	log.Debugf("Fetching from AUR (%s) successful.", url)

	var feed feed
	d := xml.NewDecoder(resp.Body)
	if err := d.Decode(&feed); err != nil {
		return nil, err
	}
	return feed.Entries, nil
}

func convert(xmlEntries []entry) []entries.Entry {
	entryList := make([]entries.Entry, len(xmlEntries))
	for i, xmlE := range xmlEntries {
		log.Debugf("Fetched entry %+v", xmlE)

		entryList[i] = entries.Entry{
			CommitTime: xmlE.convertTime(),
			Author:     xmlE.Author,
			Summary:    xmlE.Title,
		}
	}
	return entryList
}

func GetEntries(pkg string) ([]entries.Entry, error) {
	url := buildUrl(pkg)
	xmlEntries, err := fetch(url)
	if err != nil {
		return nil, err
	}
	return convert(xmlEntries), nil
}
