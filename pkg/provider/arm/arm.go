package arm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
	"github.com/Necoro/arch-log/pkg/provider"
	"github.com/Necoro/arch-log/pkg/provider/arch"
	"github.com/Necoro/arch-log/pkg/provider/aur"
)

func checkAvailability(pkg string) (string, error) {
	const URL = "https://archlinuxarm.org/data/packages/list"

	data := url.Values{}
	data.Set("search[value]", pkg)
	data.Set("start", "0")
	data.Set("length", "1")

	body := strings.NewReader(data.Encode())

	req, err := http.NewRequest(http.MethodPost, URL, body)
	if err != nil {
		return "", err
	}

	// required to make the request work
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "https://archlinuxarm.org/packages")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := provider.Request(req)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	log.Debugf("Fetching from Arch ARM (%s) successful.", URL)

	return analyzeARMData(resp)
}

type armData struct {
	Data [][]string
}

func analyzeARMData(resp io.ReadCloser) (string, error) {
	var data armData

	d := json.NewDecoder(resp)
	if err := d.Decode(&data); err != nil {
		return "", err
	}

	log.Debugf("Received ARM data: %+v", data)

	for _, d := range data.Data {
		if d[5] == "1" { // exact match
			return d[1], nil
		}
	}

	return "", entries.ErrNotFound
}

func determineBasePkg(pkg, repo string) (string, error) {
	bi, err := arch.DetermineBaseInfo(pkg, repo)
	if err == nil {
		return bi, nil
	} else if !errors.Is(err, entries.ErrNotFound) {
		return "", err
	}

	bi, err = aur.DetermineBasePkg(pkg)
	if err == nil {
		return bi, nil
	} else if !errors.Is(err, entries.ErrNotFound) {
		return "", err
	}

	log.Debugf("Package '%s' neither found on Arch nor AUR, assuming base = pkg", pkg)
	return pkg, nil
}

func setupFetch(pkg, repo string) (string, string, error) {
	if repo != "" {
		return "", "", errors.New("repo is not supported by Arch ARM")
	}

	armRepo, err := checkAvailability(pkg)
	if err != nil {
		return "", "", err
	}

	basePkg, err := determineBasePkg(pkg, armRepo)
	if err != nil {
		return "", "", err
	}

	if basePkg != pkg {
		log.Printf("Mapped pkg '%s' to pkgbase '%s'", pkg, basePkg)
	}
	return basePkg, armRepo, nil
}

func buildGHUrl(detail string) string {
	return "https://api.github.com/repos/archlinuxarm/PKGBUILDs/" + detail
}

func buildPkgBuildRequest(pkg, repo string) (*http.Request, error) {
	detail := fmt.Sprintf("contents/%s/%s/PKGBUILD", repo, pkg)
	url := buildGHUrl(detail)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// request raw data
	req.Header.Set("Accept", "application/vnd.github.raw")

	return req, nil
}

func GetPkgBuild(pkg, repo string) (io.ReadCloser, error) {
	basePkg, repo, err := setupFetch(pkg, repo)
	if err != nil {
		return nil, err
	}

	req, err := buildPkgBuildRequest(basePkg, repo)
	if err != nil {
		return nil, err
	}
	body, err := provider.Request(req)
	if err != nil {
		return nil, err
	}

	log.Debugf("Fetching from Arch ARM (%s) successful.", req.URL)

	return body, nil
}
