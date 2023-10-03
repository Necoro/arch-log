package aur

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
	"github.com/Necoro/arch-log/pkg/provider"
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

func buildUrl(pkg, verb string) string {
	return fmt.Sprintf("https://aur.archlinux.org/cgit/aur.git/%s/?h=%s", verb, pkg)
}

func fetch(url string) ([]entry, error) {
	result, err := provider.Fetch(url)
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

func convert(xmlEntries []entry) []entries.Change {
	changes := make([]entries.Change, len(xmlEntries))
	for i, xmlE := range xmlEntries {
		log.Debugf("Fetched entry %+v", xmlE)

		changes[i] = entries.Change{
			CommitTime: xmlE.convertTime(),
			Author:     xmlE.Author,
			Summary:    xmlE.Title,
			Message:    xmlE.content(),
		}
	}
	return changes
}

func setupFetch(pkg, repo string) (string, error) {
	if repo != "" {
		return "", errors.New("repo is not supported by AUR")
	}

	basePkg, err := DetermineBasePkg(pkg)
	if err != nil {
		return "", err
	}

	if basePkg != pkg {
		log.Printf("Mapped pkg '%s' to pkgbase '%s'", pkg, basePkg)
	}
	return basePkg, nil
}

func GetEntries(pkg, repo string) ([]entries.Change, error) {
	basePkg, err := setupFetch(pkg, repo)
	if err != nil {
		return nil, err
	}

	url := buildUrl(basePkg, "atom")
	xmlEntries, err := fetch(url)
	if err != nil {
		return nil, err
	}
	return convert(xmlEntries), nil
}

func GetPkgBuild(pkg, repo string) (io.ReadCloser, error) {
	basePkg, err := setupFetch(pkg, repo)
	if err != nil {
		return nil, err
	}

	url := buildUrl(basePkg, "plain/PKGBUILD")
	body, err := provider.Fetch(url)
	if err != nil {
		return nil, err
	}

	log.Debugf("Fetching from AUR (%s) successful.", url)

	return body, nil
}
