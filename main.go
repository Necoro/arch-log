package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/entries/arch"
	"github.com/Necoro/arch-log/pkg/entries/aur"
	"github.com/Necoro/arch-log/pkg/log"
)

const VERSION = "0.2.0"

// flags
var (
	printVersion bool = false
	debug        bool = false
	archForce    bool = false
	aurForce     bool = false
	reverse      bool = false
	number       int  = 10
	longLog      bool = false
)

func init() {
	flag.BoolVar(&printVersion, "version", printVersion, "print version and exit")
	flag.BoolVar(&debug, "d", debug, "enable debug output")
	flag.BoolVar(&archForce, "arch", archForce, "force usage of Arch git")
	flag.BoolVar(&aurForce, "aur", aurForce, "force usage of AUR")
	flag.BoolVar(&reverse, "r", reverse, "reverse order of commits")
	flag.IntVar(&number, "n", number, "max number of commits to show")
	flag.BoolVar(&longLog, "l", longLog, "slightly verbose log messages")
}

var timeLess = time.Time.Before

func maxTagLength(entryList []entries.Entry) int {
	maxTL := 0
	for _, e := range entryList {
		if maxTL < len(e.Tag) {
			maxTL = len(e.Tag)
		}
	}

	return maxTL
}

func formatEntryList(entryList []entries.Entry) {
	log.Debugf("Received entries: %+v", entryList)

	sort.SliceStable(entryList, func(i, j int) bool {
		return timeLess(entryList[i].CommitTime, entryList[j].CommitTime)
	})

	if len(entryList) > number {
		if reverse {
			entryList = entryList[:number]
		} else {
			entryList = entryList[len(entryList)-number:]
		}
	}

	maxTL := maxTagLength(entryList)

	for _, e := range entryList {
		if !longLog {
			fmt.Println(e.ShortFormat(maxTL))
		} else {
			fmt.Println(e.Format())
			fmt.Println("--------------")
		}
	}
}

func handleEntries(what string, pkg string, f func(string) ([]entries.Entry, error)) (bool, error) {
	log.Debug("Checking ", what)

	if entryList, err := f(pkg); err == nil {
		formatEntryList(entryList)
		return false, nil
	} else if errors.Is(err, entries.ErrNotFound) {
		log.Debug("Not found on ", what)
		return true, nil
	} else {
		return false, fmt.Errorf("error fetching from %s: %w", what, err)
	}
}

func fetch(pkg string) (notfound bool, err error) {
	if !aurForce {
		notfound, err = handleEntries("Arch", pkg, arch.GetEntries)
	}

	if aurForce || (err == nil && notfound && !archForce) {
		notfound, err = handleEntries("AUR", pkg, aur.GetEntries)
	}

	return
}

func run() error {
	flag.Parse()
	if printVersion {
		println("arch-log v. " + VERSION)
		return nil
	}

	if debug {
		log.SetDebug()
	}

	if reverse {
		timeLess = time.Time.After
	}

	if aurForce && archForce {
		log.Print("Forced both Arch and AUR, checking both.")
		aurForce = false
		archForce = false
	}

	pkg := flag.Arg(0)
	if pkg == "" {
		return fmt.Errorf("no package specified")
	}

	if nf, err := fetch(pkg); err != nil {
		return err
	} else if nf {
		var msg string
		switch {
		case aurForce:
			msg = "could not be found on AUR"
		case archForce:
			msg = "could not be found on Arch"
		default:
			msg = "could neither be found on Arch nor AUR"
		}

		return fmt.Errorf("package '%s' %s", pkg, msg)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
