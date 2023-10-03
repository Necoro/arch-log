package arch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
	"github.com/Necoro/arch-log/pkg/provider"
)

type result struct {
	PkgName string
	PkgBase string
	Repo    string
	PkgVer  string
	PkgRel  string
}

func (r result) tagName() string {
	return r.PkgVer + "-" + r.PkgRel
}

type infos struct {
	Results []result
}

// map from tag to repo
type repoInfo map[string]string

func (r repoInfo) isRestricted() bool {
	return len(r) == 1
}

func (r repoInfo) repoConstraint() string {
	if r.isRestricted() {
		for _, repo := range r {
			return repo
		}
	}

	return ""
}

func (r repoInfo) refConstraint() string {
	if r.isRestricted() {
		for tag := range r {
			return tag
		}
	}

	return "HEAD"
}

func buildPkgUrl(pkg string) string {
	return "https://archlinux.org/packages/search/json/?name=" + url.QueryEscape(pkg)
}

func fetchPkgInfo(url, repo string) (result, repoInfo, error) {
	res, err := provider.Fetch(url)
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
	}

	r := infos.Results[0]
	if len(infos.Results) == 1 && repo != "" && r.Repo != repo {
		return result{}, nil, fmt.Errorf("package '%s' only found in repo '%s', but '%s' has been requested",
			r.PkgName, r.Repo, repo)
	}

	if len(infos.Results) > 1 {
		repoInfo, err = buildRepoInfo(repo, infos.Results)
		if err != nil {
			return result{}, nil, err
		}
	}

	return r, repoInfo, nil
}

func reposString(results []result) string {
	sb := strings.Builder{}

	for i, r := range results {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteRune('\'')
		sb.WriteString(r.Repo)
		sb.WriteRune('\'')
	}

	return sb.String()
}

func buildRepoInfo(repo string, results []result) (repoInfo, error) {
	repoInfo := make(map[string]string)

	if repo == "" {
		for _, r := range results {
			repoInfo[r.tagName()] = r.Repo
		}
	} else {
		found := false
		for _, r := range results {
			if r.Repo == repo {
				repoInfo[r.tagName()] = r.Repo
				found = true
			}
		}

		if !found {
			repos := reposString(results)
			return nil, fmt.Errorf("package '%s' only found in repos %s, but '%s' has been requested",
				results[0].PkgName, repos, repo)
		}
	}
	return repoInfo, nil
}

func determineBaseInfo(pkg, repo string) (string, repoInfo, error) {
	url := buildPkgUrl(pkg)
	result, repoInfo, err := fetchPkgInfo(url, repo)

	if err != nil {
		return "", nil, err
	}

	log.Debugf("Pkg Info from Arch: %+v", result)

	if result.PkgBase != pkg {
		log.Printf("Mapped pkg '%s' to pkgbase '%s'", pkg, result.PkgBase)
	}

	return result.PkgBase, repoInfo, nil
}

func DetermineBaseInfo(pkg, repo string) (string, error) {
	bi, _, err := determineBaseInfo(pkg, repo)
	return bi, err
}
