package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/entries/arch"
	"github.com/Necoro/arch-log/pkg/entries/aur"
	"github.com/Necoro/arch-log/pkg/log"
)

func maxLength(f func(entries.Entry) string) func(entryList []entries.Entry) int {
	return func(entryList []entries.Entry) int {
		max := 0
		for _, e := range entryList {
			if max < len(f(e)) {
				max = len(f(e))
			}
		}

		return max
	}
}

var maxTagLength = maxLength(func(entry entries.Entry) string {
	return entry.Tag
})

var maxRepoLength = maxLength(func(entry entries.Entry) string {
	return entry.RepoInfo
})

func formatEntryList(entryList []entries.Entry) {
	log.Debugf("Received entries: %+v", entryList)

	sort.SliceStable(entryList, func(i, j int) bool {
		return timeLess(entryList[i].CommitTime, entryList[j].CommitTime)
	})

	if len(entryList) > options.number {
		if options.reverse {
			entryList = entryList[:options.number]
		} else {
			rest := len(entryList) - options.number
			entryList = entryList[rest:]
		}
	}

	maxTL := maxTagLength(entryList)
	maxRL := maxRepoLength(entryList)

	for _, e := range entryList {
		if !options.longLog {
			fmt.Println(e.ShortFormat(maxTL, maxRL))
		} else {
			fmt.Println(e.Format())
			fmt.Println("--------------")
		}
	}
}

func handleEntries(what string, pkg string, repo string, f func(string, string) ([]entries.Entry, error)) (bool, error) {
	log.Debug("Checking ", what)

	if entryList, err := f(pkg, repo); err == nil {
		formatEntryList(entryList)
		return false, nil
	} else if errors.Is(err, entries.ErrNotFound) {
		log.Debug("Not found on ", what)
		return true, nil
	} else {
		return false, fmt.Errorf("error fetching from %s: %w", what, err)
	}
}

func notFoundError(pkg string) error {
	var msg string
	switch {
	case options.aur:
		msg = "could not be found on AUR"
	case options.arch:
		msg = "could not be found on Arch"
	default:
		msg = "could neither be found on Arch nor AUR"
	}

	return fmt.Errorf("package '%s' %s", pkg, msg)
}

func fetchLog(pkg string) (err error) {
	var notfound bool

	if !options.aur {
		if notfound, err = handleEntries("Arch", pkg, options.repo, arch.GetEntries); err != nil {
			return
		}
	}

	if options.aur || (notfound && !options.arch) {
		if notfound, err = handleEntries("AUR", pkg, options.repo, aur.GetEntries); err != nil {
			return
		}
	}

	if notfound {
		return notFoundError(pkg)
	}

	return nil
}
