package main

import (
	"fmt"
	"sort"
	"strings"
)

const notReleaseNoted = "not release noted"

const (
	bugFix  = "bug"
	feature = "enhancement"
	other   = "other"
)

var typeToText = map[string]string{
	bugFix:  "Bug Fixes",
	feature: "Features",
	other:   "Untagged",
}

func main() {
	config := parseFlags()
	github := NewGithubClient(config.username, config.password)
	var app App
	if config.useCommits {
		app = NewCommitApp(config, github)
	} else {
		app = NewIssuesApp(config, github)
	}

	issuesByType := splitIssuesByType(app.Issues())
	if printIssues(bugFix, issuesByType[bugFix]) {
		fmt.Println("")
	}
	if printIssues(feature, issuesByType[feature]) {
		fmt.Println("")
	}
	printIssues(other, issuesByType[other])
}

func printIssues(issueType string, issues []Issue) bool {
	if len(issues) == 0 {
		return false
	}

	fmt.Printf("## %s\n", strings.ToUpper(typeToText[issueType]))

	sort.Sort(IssuesByAuthor(issues))
	for _, issue := range issues {
		fmt.Printf(" * %s [#%s](%s)\n    * Author: %s\n    * Repo: %s\n    * Release Notes: %s\n",
			issue.Title, issue.Num, issue.URL, issue.Author, issue.Repo, strings.Replace(issue.ReleaseNotes, "\n", "\n      ", -1))
	}

	return true
}

func splitIssuesByType(issues []Issue) map[string][]Issue {
	issuesByType := map[string][]Issue{}

	for _, issue := range issues {
		issueType := labelsToType(issue.Labels)
		issuesByType[issueType] = append(issuesByType[issueType], issue)
	}

	return issuesByType
}

func printSeparator(str string) {
	if len(str) == 0 {
		return
	}
	for i := 0; i < len(str); i++ {
		fmt.Printf("-")
	}
	fmt.Printf("\n")
}

func contains(list []string, v string) bool {
	for _, l := range list {
		if l == v {
			return true
		}
	}
	return false
}

func labelsToType(labels []string) string {
	if contains(labels, bugFix) {
		return bugFix
	}
	if contains(labels, feature) {
		return feature
	}
	return other
}

type App interface {
	Issues() []Issue
}

type GithubGateway interface {
	Get(path string, v interface{}) error
}

type Issue struct {
	Author       string
	Num          string
	Title        string
	Repo         string
	Labels       []string
	URL          string
	ReleaseNotes string
}
