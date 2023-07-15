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
)

func init() {
	flag.BoolVar(&printVersion, "version", printVersion, "print version and exit")
	flag.BoolVar(&debug, "d", debug, "enable debug output")
}

func formatEntryList(list []entries.Entry) {

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

	pkg := flag.Arg(0)
	if pkg == "" {
		return fmt.Errorf("no package specified")
	}

	if entryList, err := arch.GetEntries(pkg); err == nil {
		formatEntryList(entryList)
	} else if errors.Is(err, entries.ErrNotFound) {
		if entryList, err := aur.GetEntries(pkg); err == nil {
			formatEntryList(entryList)
		} else if errors.Is(err, entries.ErrNotFound) {
			return fmt.Errorf("package '%s' could neither be found on Arch nor AUR", pkg)
		} else {
			return fmt.Errorf("error fetching from AUR: %w", err)
		}
	} else {
		return fmt.Errorf("error fetching from Arch: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
