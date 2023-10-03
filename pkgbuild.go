package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/entries/arch"
	"github.com/Necoro/arch-log/pkg/log"
)

func handleResult(what string, pkg string, repo string, f func(string, string) (io.ReadCloser, error)) (bool, error) {
	log.Debug("Checking ", what)

	if body, err := f(pkg, repo); err == nil {
		err := printPkgBuild(body)
		return false, err
	} else if errors.Is(err, entries.ErrNotFound) {
		log.Debug("Not found on ", what)
		return true, nil
	} else {
		return false, fmt.Errorf("error fetching from %s: %w", what, err)
	}
}

func fetchPkgBuild(pkg string) (err error) {
	var notfound bool

	if !options.aur {
		if notfound, err = handleResult("Arch", pkg, options.repo, arch.GetPkgBuild); err != nil {
			return
		}
	}

	//if options.aur || (notfound && !options.arch) {
	//	if notfound, err = handleResult("AUR", pkg, options.repo, aur.GetEntries); err != nil {
	//		return
	//	}
	//}

	if notfound {
		return notFoundError(pkg)
	}

	return nil
}

func printPkgBuild(body io.ReadCloser) error {
	defer body.Close()

	if _, err := io.Copy(os.Stdout, body); err != nil {
		return fmt.Errorf("writing PKGBUILD: %v", err)
	}

	return nil
}
