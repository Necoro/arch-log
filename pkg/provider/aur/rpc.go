package aur

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
	"github.com/Necoro/arch-log/pkg/provider"
)

type result struct {
	Name        string
	PackageBase string
}

type infos struct {
	Results []result
}

func buildRpcUrl(pkg string) string {
	return "https://aur.archlinux.org/rpc/?v=5&type=info&arg[]=" + url.QueryEscape(pkg)
}

func fetchRpcInfo(url string) (result, error) {
	res, err := provider.Fetch(url)
	if err != nil {
		return result{}, err
	}
	defer res.Close()

	log.Debugf("Fetching from AUR RPC (%s) successful.", url)

	var infos infos
	d := json.NewDecoder(res)
	if err = d.Decode(&infos); err != nil {
		return result{}, err
	}

	if len(infos.Results) == 0 {
		return result{}, entries.ErrNotFound
	} else if len(infos.Results) > 1 {
		return result{}, fmt.Errorf("more than one package info found: %+v", infos.Results)
	}

	return infos.Results[0], nil
}

func DetermineBasePkg(pkg string) (string, error) {
	url := buildRpcUrl(pkg)
	result, err := fetchRpcInfo(url)

	if err != nil {
		return "", err
	}
	log.Debugf("Pkg Info from AUR RPC: %+v", result)

	return result.PackageBase, nil
}
