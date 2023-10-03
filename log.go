package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
	"github.com/Necoro/arch-log/pkg/provider/arch"
	"github.com/Necoro/arch-log/pkg/provider/aur"
)

func maxLength(f func(entries.Change) string) func(changes []entries.Change) int {
	return func(changes []entries.Change) int {
		max := 0
		for _, e := range changes {
			if max < len(f(e)) {
				max = len(f(e))
			}
		}

		return max
	}
}

var maxTagLength = maxLength(func(change entries.Change) string {
	return change.Tag
})

var maxRepoLength = maxLength(func(change entries.Change) string {
	return change.RepoInfo
})

func formatEntryList(changes []entries.Change) {
	log.Debugf("Received entries: %+v", changes)

	sort.SliceStable(changes, func(i, j int) bool {
		return timeLess(changes[i].CommitTime, changes[j].CommitTime)
	})

	if len(changes) > options.number {
		if options.reverse {
			changes = changes[:options.number]
		} else {
			rest := len(changes) - options.number
			changes = changes[rest:]
		}
	}

	maxTL := maxTagLength(changes)
	maxRL := maxRepoLength(changes)

	for _, c := range changes {
		if !options.longLog {
			fmt.Println(c.ShortFormat(maxTL, maxRL))
		} else {
			fmt.Println(c.Format())
			fmt.Println("--------------")
		}
	}
}

func handleEntries(what, pkg string, repo string, f func(string, string) ([]entries.Change, error)) (bool, error) {
	log.Debug("Checking ", what)

	if changes, err := f(pkg, repo); err == nil {
		formatEntryList(changes)
		return true, nil
	} else if errors.Is(err, entries.ErrNotFound) {
		log.Debug("Not found on ", what)
		return false, nil
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
	case options.armOnly:
		msg = "could not be found on Arch ARM"
	default:
		msg = "could neither be found on Arch nor AUR"
	}

	if options.arm {
		msg += " nor Arch ARM"
	}

	return fmt.Errorf("package '%s' %s", pkg, msg)
}

func fetchLog(pkg string) error {
	if options.arm || options.armOnly {
		if done, err := handleEntries("Arch ARM", pkg, options.repo, arch.GetEntries); err != nil {
			return err
		} else if !done && options.armOnly {
			return notFoundError(pkg)
		} else if done {
			return nil
		}
	}

	if !options.aur {
		if done, err := handleEntries("Arch", pkg, options.repo, arch.GetEntries); err != nil {
			return err
		} else if !done && options.arch {
			return notFoundError(pkg)
		} else if done {
			return nil
		}
	}

	if done, err := handleEntries("AUR", pkg, options.repo, aur.GetEntries); err != nil {
		return err
	} else if !done {
		return notFoundError(pkg)
	}

	return nil
}
