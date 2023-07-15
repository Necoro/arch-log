package main

import (
	"flag"
	"os"

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

func run() error {
	flag.Parse()
	if printVersion {
		println("arch-log v. devel")
		return nil
	}

	if debug {
		log.SetDebug()
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
