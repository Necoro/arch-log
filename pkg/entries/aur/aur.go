package aur

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/http"
	"github.com/Necoro/arch-log/pkg/log"
)

type feed struct {
	Entries []entry `xml:"entry"`
}

type content struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

type entry struct {
	Title   string    `xml:"title"`
	Updated string    `xml:"updated"`
	Author  string    `xml:"author>name"`
	Content []content `xml:"content"`
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

func (e entry) content() string {
	for _, c := range e.Content {
		if c.Type == "text" {
			return strings.TrimSpace(c.Text)
		}
	}

	return ""
}

func buildUrl(pkg string) string {
	return "https://aur.archlinux.org/cgit/aur.git/atom/?h=" + pkg
}

func fetch(url string) ([]entry, error) {
	result, err := http.Fetch(url)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	log.Debugf("Fetching from AUR (%s) successful.", url)

	var feed feed
	d := xml.NewDecoder(result)
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
			Message:    xmlE.content(),
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
