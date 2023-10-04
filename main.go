package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/Necoro/arch-log/pkg/log"
)

//go:embed VERSION
var VERSION string

const PROG_NAME = "arch-log"

var versionMsg = PROG_NAME + " v" + VERSION

// flags
var options struct {
	printVersion bool
	debug        bool
	arch         bool
	aur          bool
	repo         string
	pkgbuild     bool
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
	flag.BoolVarP(&options.pkgbuild, "pkgbuild", "p", false, "show PKGBUILD instead of the log (honors PAGER)")
}

var timeLess = time.Time.Before

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
	} else if options.repo != "" {
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

	if options.pkgbuild {
		log.Debug("Showing PKGBUILD instead of log")
		return fetchPkgBuild(pkg)
	}

	return fetchLog(pkg)
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
