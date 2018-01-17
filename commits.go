package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type CommitApp struct {
	config Config
	github GithubGateway
}

func NewCommitApp(config Config, github GithubGateway) CommitApp {
	return CommitApp{config, github}
}

func (app CommitApp) Issues() []Issue {
	issues := []Issue{}

	for _, repo := range strings.Fields(app.config.repos) {
		issues = append(issues, app.repoIssues(repo)...)
	}

	return issues
}

func (app CommitApp) repoIssues(repo string) []Issue {
	prs := []Issue{}

	var res struct {
		Commits []struct {
			Commit struct {
				Author struct {
					Name string
				}
				Message string
			}
		}
	}

	err := app.github.Get(fmt.Sprintf("repos/%v/%v/compare/%v...%v",
		app.config.owner,
		repo,
		app.config.base,
		app.config.head), &res)

	if err != nil {
		log.Fatalf("%v", err)
	}

	re1 := regexp.MustCompile("Merge pull request #(?P<num>\\d+).*\n\n(?P<title>.*)")
	re2 := regexp.MustCompile("(?P<title>.*) \\(#(?P<num>\\d+)\\)")

	for _, commit := range res.Commits {
		for _, re := range []*regexp.Regexp{re1, re2} {
			matches := re.FindStringSubmatch(commit.Commit.Message)
			if matches != nil && len(matches) == 3 {
				result := make(map[string]string)
				for i, name := range re.SubexpNames() {
					if i != 0 {
						result[name] = matches[i]
					}
				}

				prs = append(prs, Issue{
					Author: commit.Commit.Author.Name,
					Repo:   repo,
					Num:    result["num"],
					URL:    fmt.Sprintf("https://github.com/%s/%s/pull/%s", app.config.owner, repo, result["num"]),
					Title:  result["title"],
				})

				break
			}
		}
	}

	return prs
}
