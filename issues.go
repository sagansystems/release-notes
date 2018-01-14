package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

type IssuesApp struct {
	config Config
	github GithubGateway
}

func NewIssuesApp(config Config, github GithubGateway) IssuesApp {
	return IssuesApp{config, github}
}

func (app IssuesApp) Issues() []Issue {
	issues := []Issue{}

	for _, repo := range strings.Fields(app.config.repos) {
		issues = append(issues, app.repoIssues(repo)...)
	}

	return issues
}

type Item struct {
	Title       string `json:"title"`
	ClosedAt    string `json:"closed_at"`
	UpdatedAt   string `json:"updated_at"`
	Number      int    `json:"number"`
	Body        string `json:"body"`
	PullRequest struct {
		URL string `json:"html_url"`
	} `json:"pull_request"`
	User struct {
		Login string `json:"login"`
	} `json:"user`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func (i Item) toIssue(repo string) Issue {
	var labels []string
	for _, label := range i.Labels {
		labels = append(labels, label.Name)
	}
	return Issue{
		Author:       i.User.Login,
		Repo:         repo,
		Num:          fmt.Sprintf("%d", i.Number),
		URL:          i.PullRequest.URL,
		Title:        i.Title,
		Labels:       labels,
		ReleaseNotes: strings.TrimSpace(getReleaseNotes(i.Body)),
	}
}

type Items []Item

func (app IssuesApp) repoIssues(repo string) []Issue {
	issues := []Issue{}

	var res struct {
		Items Items `json:"items"`
	}

	var res2 struct {
		Items Items `json:"items"`
	}

	since := releaseToTimestamp(app.config.base)
	until := releaseToTimestamp(app.config.head)

	url := fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:merged+closed:%s..%s+no:milestone", repo, since, until)
	if err := app.github.Get(url, &res); err != nil {
		fmt.Printf("error fetching issues without milestones: %v", err)
		os.Exit(1)
	}

	url = fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:merged+milestone:%s", repo, app.config.head)
	if err := app.github.Get(url, &res2); err != nil {
		fmt.Printf("error fetching issues with milestones: %v", err)
		os.Exit(1)
	}
	res.Items = append(res.Items, res2.Items...)

	for _, item := range res.Items {
		issues = append(issues, item.toIssue(repo))
	}

	return issues
}

const tsRegexp = ".*-(\\d{4})(\\d{2})(\\d{2})(\\d{2})(\\d{2})(\\d{2})Z"
const timeFmt = "%s-%s-%sT%s:%s:%sZ"

func releaseToTimestamp(release string) string {
	re := regexp.MustCompile(tsRegexp)
	matches := re.FindStringSubmatch(release)
	if len(matches) < 7 {
		return time.Now().UTC().Format(fmt.Sprintf(timeFmt, "2006", "01", "02", "15", "04", "05"))
	}
	return fmt.Sprintf(timeFmt, matches[1], matches[2], matches[3], matches[4], matches[5], matches[6])
}

func getReleaseNotes(description string) string {
	const notesRegexp = "(?:##+\\s*|\\*\\*)(?i:release\\s*notes)\\**[\\r\\n]+((?s:.)*?)(?:\\z|\\*\\*|##+)"
	re := regexp.MustCompile(notesRegexp)
	matches := re.FindStringSubmatch(description)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimRight(strings.Replace(matches[1], "\r", "", -1), "\n\r ")
}
