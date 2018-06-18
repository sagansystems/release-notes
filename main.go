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
	labelDemo            = Label("demo")
)

type Category string

const (
	categoryBugFix        = Category("bug")
	categoryDemo          = Category("demo")
	categoryEnhancement   = Category("enhancement")
	categoryInternal      = Category("internal")
	categoryOmitted       = Category("omitted")
	categoryPendingHotfix = Category("pendingHotfix")
	categoryUntagged      = Category("untagged")
	categoryUpcoming      = Category("upcoming")
)

var categoryToText = map[Category]string{
	categoryEnhancement:   "Enhancements",
	categoryBugFix:        "Bug Fixes",
	categoryUntagged:      "Untagged",
	categoryPendingHotfix: "Pending Hotfixes",
	categoryUpcoming:      "Upcoming",
	categoryInternal:      "Internal",
	categoryDemo:          "Demo Changes",
}

var outputCategories = []Category{
	categoryEnhancement,
	categoryBugFix,
	categoryUntagged,
	categoryPendingHotfix,
	categoryUpcoming,
	categoryInternal,
	categoryDemo,
}

func main() {
	config := parseFlags()
	githubDirect := NewGithubClient(config.username, config.password)
	goGithub := oauth2Client(config.password)
	parseBranchTemplate(config, goGithub)

	var app App
	if config.useCommits {
		app = NewCommitApp(config, githubDirect)
	} else {
		app = NewIssuesApp(config, githubDirect)
	}

	filteredIssues := filterIssues(app.Issues(), config.hotfixOnly, config.withInternal)
	issuesByCategory := splitIssuesByCategory(filteredIssues)

	fmt.Println("<!DOCTYPE html>")
	fmt.Println("<html>")
	fmt.Println("<body>")
	for _, category := range outputCategories {
		printIssues(categoryToText[category], issuesByCategory[category], config.withReleaseNotes, config.withTesting)
	}
	fmt.Println("</body>")
	fmt.Println("</html>")
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

		fmt.Printf("<b>%s</b>&nbsp;-&nbsp;<a href=\"%s\">#%s</a>&nbsp;-&nbsp;%s<br/>%s%s<br/>\n",
			issue.Title, issue.URL, issue.Num, issue.Author, releaseNotes, testing)
	}

	return true
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
	isDemoChange := contains(issue.Labels, labelDemo)
	isEnhancement := contains(issue.Labels, labelEnhancement)
	isReleaseNoted := !contains(issue.Labels, labelNotReleaseNoted)

	if isDemoChange {
		return categoryDemo
	}

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

func nlToBr(str string) string {
	return strings.Replace(str, "\n", "<br>", -1)
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
