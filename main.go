package main

import (
	"fmt"
	"sort"
	"strings"
)

type Label string

const (
	labelBug             = Label("bug")
	labelEnhancement     = Label("enhancement")
	labelNotReleaseNoted = Label("not release noted")
)

type Category string

const (
	categoryBugFix        = Category("bug")
	categoryEnhancement   = Category("enhancement")
	categoryUntagged      = Category("untagged")
	categoryPendingHotfix = Category("pendingHotfix")
	categoryUpcoming      = Category("upcoming")
	categoryInternal      = Category("internal")
	categoryOmitted       = Category("omitted")
)

var categoryToText = map[Category]string{
	categoryEnhancement:   "Enhancements",
	categoryBugFix:        "Bug Fixes",
	categoryUntagged:      "Untagged",
	categoryPendingHotfix: "Pending Hotfixes",
	categoryUpcoming:      "Upcoming",
	categoryInternal:      "Internal",
}

var outputCategories = []Category{
	categoryEnhancement,
	categoryBugFix,
	categoryUntagged,
	categoryPendingHotfix,
	categoryUpcoming,
	categoryInternal,
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
	issuesByCategory := splitIssuesByCategory(filteredIssues)
	for _, category := range outputCategories {
		printIssues(categoryToText[category], issuesByCategory[category], config.withReleaseNotes, config.withTesting)
	}
}

func printIssues(label string, issues []Issue, withReleaseNotes, withTesting bool) bool {
	if len(issues) == 0 {
		return false
	}

	sort.Sort(IssuesByAuthorAndInternal(issues))

	if label != "" {
		fmt.Printf("<h3>%s</h3>\n", label)
	}

	for _, issue := range issues {
		releaseNotes := ""
		if withReleaseNotes && issue.ReleaseNotes != "" {
			releaseNotes = nlToBr(issue.ReleaseNotes) + "<br/>"
		}
		testing := ""
		if withTesting && issue.Testing != "" {
			testing = nlToBr(issue.Testing) + "<br/>"
		}
		hotfix := ""
		if issue.IsHotfix {
			hotfix = "HOTFIX: "
		}
		issueNotReleaseNoted := ""
		if contains(issue.Labels, labelNotReleaseNoted) {
			issueNotReleaseNoted = "INTERNAL: "
		}

		fmt.Printf("<b>%s%s%s</b>&nbsp;-&nbsp;<a href=\"%s\">#%s</a>&nbsp;-&nbsp;%s<br/>%s%s<br/>\n",
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
		releaseNoted := !contains(issue.Labels, labelNotReleaseNoted)
		if (!hotfixOnly || issue.IsHotfix) && (releaseNoted || withInternal) {
			filteredIssues = append(filteredIssues, issue)
		}
	}

	return filteredIssues
}

func splitIssuesByCategory(issues []Issue) map[Category][]Issue {
	issuesByCategory := map[Category][]Issue{}

	for _, issue := range issues {
		category := issueCategory(issue)
		issuesByCategory[category] = append(issuesByCategory[category], issue)
	}

	return issuesByCategory
}

func issueCategory(issue Issue) Category {
	isBugFix := contains(issue.Labels, labelBug)
	isEnhancement := contains(issue.Labels, labelEnhancement)
	isReleaseNoted := !contains(issue.Labels, labelNotReleaseNoted)

	if issue.IsMerged {
		if contains(issue.Labels, labelNotReleaseNoted) {
			return categoryInternal
		}
		if isBugFix {
			return categoryBugFix
		}
		if isEnhancement {
			return categoryEnhancement
		}
		return categoryUntagged
	}
	if issue.IsHotfix {
		return categoryPendingHotfix
	}
	if isReleaseNoted {
		return categoryUpcoming
	}
	return categoryOmitted
}

func contains(list []string, v Label) bool {
	for _, l := range list {
		if l == string(v) {
			return true
		}
	}
	return false
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
