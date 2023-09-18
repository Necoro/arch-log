package arch

import (
	"encoding/json"
	"net/url"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/http"
	"github.com/Necoro/arch-log/pkg/log"
)

type result struct {
	PkgName string
	PkgBase string
	Repo    string
	PkgVer  string
	PkgRel  string
}

type infos struct {
	Results []result
}

type repoInfo map[string]string

func buildPkgUrl(pkg string) string {
	return "https://archlinux.org/packages/search/json/?name=" + url.QueryEscape(pkg)
}

func fetchPkgInfo(url string) (result, repoInfo, error) {
	res, err := http.Fetch(url)
	if err != nil {
		return result{}, nil, err
	}
	defer res.Close()

	log.Debugf("Fetching from Arch PkgInfo (%s) successful.", url)

	var infos infos
	d := json.NewDecoder(res)
	if err = d.Decode(&infos); err != nil {
		return result{}, nil, err
	}

	var repoInfo repoInfo
	if len(infos.Results) == 0 {
		return result{}, repoInfo, entries.ErrNotFound
	} else if len(infos.Results) > 1 {
		repoInfo = make(map[string]string)
		for _, r := range infos.Results {
			tagName := r.PkgVer + "-" + r.PkgRel
			repoInfo[tagName] = r.Repo
		}
	}

	return infos.Results[0], repoInfo, nil
}

func determineBaseInfo(pkg string) (string, repoInfo, error) {
	url := buildPkgUrl(pkg)
	result, repoInfo, err := fetchPkgInfo(url)

	if err != nil {
		return "", nil, err
	}
	log.Debugf("Pkg Info from Arch: %+v", result)

	return result.PkgBase, repoInfo, nil
}
