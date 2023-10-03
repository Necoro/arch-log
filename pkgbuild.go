package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/entries/arch"
	"github.com/Necoro/arch-log/pkg/entries/aur"
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

	if options.aur || (notfound && !options.arch) {
		if notfound, err = handleResult("AUR", pkg, options.repo, aur.GetPkgBuild); err != nil {
			return
		}
	}

	if notfound {
		return notFoundError(pkg)
	}

	return nil
}

func printPkgBuild(body io.ReadCloser) error {
	defer body.Close()

	if pager := os.Getenv("PAGER"); pager != "" {
		log.Debugf("'PAGER' set as '%s'", pager)
		return writeToPager(body, pager)
	}

	if _, err := io.Copy(os.Stdout, body); err != nil {
		return fmt.Errorf("writing PKGBUILD: %v", err)
	}

	return nil
}

func writeToPager(body io.ReadCloser, pager string) error {
	args := strings.Split(pager, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer pipe.Close()
		io.Copy(pipe, body)
	}()

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("running PAGER '%s': %v", pager, err)
	}

	return nil
}
