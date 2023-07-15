package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/entries/arch"
	"github.com/Necoro/arch-log/pkg/entries/aur"
	"github.com/Necoro/arch-log/pkg/log"
)

// flags
var (
	printVersion bool = false
	debug        bool = false
	archForce    bool = false
	aurForce     bool = false
)

func init() {
	flag.BoolVar(&printVersion, "version", printVersion, "print version and exit")
	flag.BoolVar(&debug, "d", debug, "enable debug output")
	flag.BoolVar(&archForce, "arch", archForce, "force usage of Arch git")
	flag.BoolVar(&aurForce, "aur", aurForce, "force usage of AUR")
}

func formatEntryList(list []entries.Entry) {
	log.Debugf("Received entries: %+v", list)
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
		println("arch-log v. devel")
		return nil
	}

	if debug {
		log.SetDebug()
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
