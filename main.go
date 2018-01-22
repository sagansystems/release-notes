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

	filteredIssues := filterIssues(app.Issues(), config.hotfixOnly, config.withInternal)
	issuesByType := splitIssuesByType(filteredIssues)
	if printIssues(bugFix, issuesByType[bugFix], config.withReleaseNotes, config.withTesting) {
		fmt.Println("")
	}
	if printIssues(feature, issuesByType[feature], config.withReleaseNotes, config.withTesting) {
		fmt.Println("")
	}
	printIssues(other, issuesByType[other], config.withReleaseNotes, config.withTesting)
}

func printIssues(issueType string, issues []Issue, withReleaseNotes, withTesting bool) bool {
	if len(issues) == 0 {
		return false
	}

	sort.Sort(IssuesByAuthor(issues))

	fmt.Printf("<h2>%s</h2>\n", strings.ToUpper(typeToText[issueType]))
	for _, issue := range issues {
		releaseNotes := ""
		if withReleaseNotes && issue.ReleaseNotes != "" {
			releaseNotes = "<i>Release Notes:</i><br/><pre>" + issue.ReleaseNotes + "</pre><br/>"
		}
		testing := ""
		if withTesting && issue.Testing != "" {
			testing = "<i>Testing:</i><br/><pre>" + issue.Testing + "</pre><br/>"
		}
		hotfix := ""
		if issue.IsHotfix {
			hotfix = "HOTFIX: "
		}
		issueNotReleaseNoted := ""
		if contains(issue.Labels, notReleaseNoted) {
			issueNotReleaseNoted = "NOT RELEASE NOTED: "
		}

		fmt.Printf("<b>%s%s%s</b><br/><a href=\"%s\">#%s</a>&nbsp;-&nbsp;%s<br/>%s%s<br/>\n",
			hotfix, issueNotReleaseNoted, issue.Title, issue.URL, issue.Num, issue.Author, releaseNotes, testing)
	}

	return true
}

func filterIssues(issues []Issue, hotfixOnly, withInternal bool) []Issue {
	var filteredIssues []Issue

	for _, issue := range issues {
		releaseNoted := !contains(issue.Labels, notReleaseNoted)
		if (!hotfixOnly || issue.IsHotfix) && (releaseNoted || withInternal) {
			filteredIssues = append(filteredIssues, issue)
		}
	}

	return filteredIssues
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
	Testing      string
	IsHotfix     bool
}
