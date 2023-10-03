package arch

import (
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/Necoro/arch-log/pkg/entries"
	"github.com/Necoro/arch-log/pkg/log"
	"github.com/Necoro/arch-log/pkg/provider"
)

type commit struct {
	Title     string
	Timestamp string `json:"created_at"`
	Author    string `json:"author_name"`
	Message   string
	Id        string
}

type tag struct {
	Name   string
	Commit struct{ Id string }
}

func (e commit) convertTime() time.Time {
	if e.Timestamp == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
		log.Warnf("Problem parsing time '%s' -- ignoring: %v.", e.Timestamp, err)
		return time.Time{}
	} else {
		return t
	}
}

func (e commit) cleanedMessage() string {
	if e.Message == e.Title {
		return ""
	}

	headerEnd := strings.Index(e.Message, "\n\n")
	if headerEnd == -1 {
		return e.Message
	}
	return e.Message[headerEnd+1:]
}

func fetch(url string, jsonEntries any) error {
	result, err := provider.Fetch(url)
	if err != nil {
		return err
	}
	defer result.Close()

	log.Debugf("Fetching from Arch (%s) successful.", url)

	d := json.NewDecoder(result)
	if err := d.Decode(jsonEntries); err != nil {
		return err
	}
	return nil
}

func buildCommitsUrl(pkg string) string {
	return buildUrl(pkg, "commits")
}

func buildTagsUrl(pkg string) string {
	return buildUrl(pkg, "tags")
}

func buildPkgBuildUrl(pkg, ref string) string {
	filePath := "files/PKGBUILD/raw?ref=" + url.QueryEscape(ref)

	return buildUrl(pkg, filePath)
}

func buildUrl(pkg, action string) string {
	repoName := url.QueryEscape("archlinux/packaging/packages/" + pkg)

	return "https://gitlab.archlinux.org/api/v4/projects/" + repoName + "/repository/" + action
}

func groupTag(tags []tag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[t.Commit.Id] = t.Name
	}

	return m
}

func convert(commits []commit, tags []tag, repoInfo repoInfo) []entries.Change {
	changeList := make([]entries.Change, 0, len(commits))
	tagMap := groupTag(tags)
	constrain := repoInfo.isRestricted()
	constrainRepo := repoInfo.repoConstraint()
	printRepo := !constrain

	if constrain {
		log.Printf("Restricting commits to repo '%s'", constrainRepo)
	}

	for _, c := range commits {
		log.Debugf("Fetched commit %+v", c)

		tag := tagMap[c.Id]
		repo := repoInfo[tag]

		if !constrain || constrainRepo == repo {
			constrain = false

			c := entries.Change{
				CommitTime: c.convertTime(),
				Author:     c.Author,
				Summary:    c.Title,
				Message:    c.cleanedMessage(),
				Tag:        tagMap[c.Id]}

			if printRepo {
				c.RepoInfo = repo
			}

			changeList = append(changeList, c)
		}
	}
	return changeList
}

//goland:noinspection GoImportUsedAsName
func GetEntries(pkg, repo string) ([]entries.Change, error) {
	basePkg, repoInfo, err := determineBaseInfo(pkg, repo)
	if err != nil {
		return nil, err
	}

	url := buildCommitsUrl(basePkg)
	var commits []commit
	if err := fetch(url, &commits); err != nil {
		return nil, err
	}

	url = buildTagsUrl(basePkg)
	var tags []tag
	if err := fetch(url, &tags); err != nil {
		return nil, err
	}

	return convert(commits, tags, repoInfo), nil
}

func GetPkgBuild(pkg, repo string) (io.ReadCloser, error) {
	basePkg, repoInfo, err := determineBaseInfo(pkg, repo)
	if err != nil {
		return nil, err
	}

	commitRef := repoInfo.refConstraint()

	url := buildPkgBuildUrl(basePkg, commitRef)
	body, err := provider.Fetch(url)
	if err != nil {
		body.Close()
		return nil, err
	}

	log.Debugf("Fetching from Arch (%s) successful.", url)

	return body, nil
}
