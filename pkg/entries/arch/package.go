package arch

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Necoro/arch-log/pkg/http"
	"github.com/Necoro/arch-log/pkg/log"
)

type result struct {
	PkgName string
	PkgBase string
	Repo    string
}

type infos struct {
	Results []result
}

func buildPkgUrl(pkg string) string {
	return "https://archlinux.org/packages/search/json/?name=" + url.QueryEscape(pkg)
}

func fetchPkgInfo(url string) (result, error) {
	res, err := http.Fetch(url)
	if err != nil {
		return result{}, err
	}
	defer res.Close()

	log.Debugf("Fetching from Arch PkgInfo (%s) successful.", url)

	var infos infos
	d := json.NewDecoder(res)
	if err = d.Decode(&infos); err != nil {
		return result{}, err
	}

	if len(infos.Results) > 1 {
		return result{}, fmt.Errorf("more than one package info found: %+v", infos.Results)
	}

	return infos.Results[0], nil
}

func determineBasePkg(pkg string) (string, error) {
	url := buildPkgUrl(pkg)
	result, err := fetchPkgInfo(url)

	if err != nil {
		return "", err
	}
	log.Debugf("Pkg Info from Arch: %+v", result)

	return result.PkgBase, nil
}
