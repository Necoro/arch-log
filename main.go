package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/entries/arch"
	"github.com/Necoro/arch-log/pkg/entries/aur"
	"github.com/Necoro/arch-log/pkg/log"
)

const VERSION = "0.2.1"
const PROG_NAME = "arch-log"

var versionMsg = PROG_NAME + " v" + VERSION

// flags
var options struct {
	printVersion bool
	debug        bool
	arch         bool
	aur          bool
	repo         string
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
	flag.StringVar(&options.repo, "repo", "", "restrict to repo (e.g. \"extra\")")
}

var timeLess = time.Time.Before

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

func fetch(pkg string) (err error) {
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

func parseFlags() (string, error) {
	// overwrite errorHandling mode
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			log.Errorf("%v\n\n", err)
			flag.Usage()
		}

		return "", nil
	}

	if options.printVersion {
		println(versionMsg)
		return "", nil
	}

	if options.debug {
		log.SetDebug()
	}

	if options.reverse {
		timeLess = time.Time.After
	}

	pkg := flag.Arg(0)
	if pkg == "" {
		return "", errors.New("no package specified")
	}

	if idx := strings.IndexRune(pkg, '/'); idx > -1 {
		repo := pkg[:idx]
		pkg = pkg[idx+1:]

		log.Debugf("Split package name into repo '%s' and pkg '%s'.", repo, pkg)

		if options.repo == "" {
			options.repo = repo
		} else if options.repo != repo {
			return "", fmt.Errorf("conflicting repos specified: '%s' vs '%s'", options.repo, repo)
		}
	}

	if strings.ToLower(options.repo) == "aur" {
		log.Debug("Found repo 'AUR', assuming '--aur'")
		options.aur = true
		options.repo = ""
	} else {
		log.Debug("Repo is given, assuming '--arch'")
		options.arch = true
	}

	if options.aur && options.arch {
		log.Print("Forced both Arch and AUR, checking both.")
		options.aur = false
		options.arch = false
	}

	return pkg, nil
}

func run() error {
	pkg, err := parseFlags()
	if err != nil || pkg == "" {
		return err
	}

	return fetch(pkg)
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
