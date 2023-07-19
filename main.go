package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/entries/arch"
	"github.com/Necoro/arch-log/pkg/entries/aur"
	"github.com/Necoro/arch-log/pkg/log"
)

const VERSION = "0.2.0"
const PROG_NAME = "arch-log"

var versionMsg = errors.New(PROG_NAME + " v." + VERSION)

// flags
var options struct {
	printVersion bool
	debug        bool
	arch         bool
	aur          bool
	reverse      bool
	number       int
	longLog      bool
}

func init() {
	flag.BoolVar(&options.printVersion, "version", false, "print version and exit")
	flag.BoolVarP(&options.debug, "debug", "d", false, "enable debug output")
	flag.BoolVar(&options.arch, "arch", false, "force usage of Arch git")
	flag.BoolVar(&options.aur, "aur", false, "force usage of AUR")
	flag.BoolVarP(&options.reverse, "reverse", "r", false, "reverse order of commits")
	flag.IntVarP(&options.number, "number", "n", 10, "max number of commits to show")
	flag.BoolVarP(&options.longLog, "long", "l", false, "slightly verbose log messages")
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

	if len(entryList) > options.number {
		if options.reverse {
			entryList = entryList[:options.number]
		} else {
			rest := len(entryList) - options.number
			entryList = entryList[rest:]
		}
	}

	maxTL := maxTagLength(entryList)

	for _, e := range entryList {
		if !options.longLog {
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

	if !options.aur {
		notfound, err = handleEntries("Arch", pkg, arch.GetEntries)
	}

	if options.aur || (err == nil && notfound && !options.arch) {
		notfound, err = handleEntries("AUR", pkg, aur.GetEntries)
	}

	return
}

func parseFlags() (string, error) {
	flag.Parse()

	if options.printVersion {
		return "", versionMsg
	}

	if options.debug {
		log.SetDebug()
	}

	if options.reverse {
		timeLess = time.Time.After
	}

	if options.aur && options.arch {
		log.Print("Forced both Arch and AUR, checking both.")
		options.aur = false
		options.arch = false
	}

	pkg := flag.Arg(0)
	if pkg == "" {
		return "", fmt.Errorf("no package specified")
	}
	return pkg, nil
}

func run() error {
	pkg, err := parseFlags()
	if err != nil {
		if errors.Is(err, versionMsg) {
			return nil
		}
		return err
	}

	if nf, err := fetch(pkg); err != nil {
		return err
	} else if nf {
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

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
