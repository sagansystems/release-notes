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

const (
	merged        = ""
	pendingHotfix = "Pending Hotfix"
	upcoming      = "Future"
)

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

	sort.Sort(IssuesByAuthorAndInternal(issues))

	fmt.Printf("<h2>%s</h2>\n", strings.ToUpper(typeToText[issueType]))

	issuesByCategory := splitIssuesByCategory(issues)

	printIssuesCategory(merged, issuesByCategory[merged], withReleaseNotes, withTesting)
	if printIssuesCategory(pendingHotfix, issuesByCategory[pendingHotfix], withReleaseNotes, withTesting) {
		fmt.Println("")
	}
	filteredPending := filterIssues(issuesByCategory[upcoming], false, false)
	if printIssuesCategory(upcoming, filteredPending, withReleaseNotes, withTesting) {
		fmt.Println("")
	}

	return true
}

func printIssuesCategory(category string, issues []Issue, withReleaseNotes, withTesting bool) bool {
	if len(issues) == 0 {
		return false
	}

	if category != "" {
		fmt.Printf("<h3>%s</h3>\n", category)
	}

	for _, issue := range issues {
		releaseNotes := ""
		if withReleaseNotes && issue.ReleaseNotes != "" {
			releaseNotes = "<i>Release Notes:</i><br/>" + nlToBr(issue.ReleaseNotes) + "<br/>"
		}
		testing := ""
		if withTesting && issue.Testing != "" {
			testing = "<i>Testing:</i><br/>" + nlToBr(issue.Testing) + "<br/>"
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

func nlToBr(str string) string {
	return strings.Replace(str, "\n", "<br>", -1)
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

func splitIssuesByCategory(issues []Issue) map[string][]Issue {
	issuesByCategory := map[string][]Issue{}

	for _, issue := range issues {
		category := issueCategory(issue)
		issuesByCategory[category] = append(issuesByCategory[category], issue)
	}

	return issuesByCategory
}

func issueCategory(issue Issue) string {
	if issue.IsHotfix && !issue.IsMerged {
		return pendingHotfix
	}
	if issue.IsMerged {
		return merged
	}
	return upcoming
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
	IsMerged     bool
}
