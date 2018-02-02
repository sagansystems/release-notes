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

	secondsPerRepo := 9
	repos := strings.Fields(app.config.repos)
	numRepos := len(repos)
	shouldThrottle := numRepos > 7
	timeRemaining := numRepos * secondsPerRepo
	if numRepos > 7 {
		fmt.Fprintf(os.Stderr, "processing: approximately ")
		fmt.Fprintf(os.Stderr, "\033[s")
		fmt.Fprintf(os.Stderr, "%d seconds remaining", timeRemaining)
	}
	for _, repo := range repos {
		start := time.Now()
		issues = append(issues, app.repoIssues(repo)...)
		elapsed := int(time.Since(start).Round(time.Second).Seconds())
		if shouldThrottle {
			timeRemaining = processingWait(secondsPerRepo-elapsed, timeRemaining-elapsed)
		}
	}
	if shouldThrottle {
		fmt.Fprintf(os.Stderr, "\033[1E")
	}

	return issues
}

func processingWait(seconds, timeRemaining int) int {
	for i := 0; i < seconds; i++ {
		fmt.Fprintf(os.Stderr, "\033[u")
		fmt.Fprintf(os.Stderr, "\033[s")
		fmt.Fprintf(os.Stderr, "%d", timeRemaining-i)
		time.Sleep(time.Second)
	}
	return timeRemaining - seconds
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

func (i Item) toIssue(repo string, isHotfix bool, isMerged bool) Issue {
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
		Testing:      strings.TrimSpace(getTesting(i.Body)),
		IsHotfix:     isHotfix,
		IsMerged:     isMerged,
	}
}

type Items []Item

func (app IssuesApp) repoIssues(repo string) []Issue {
	issues := []Issue{}

	type Response struct {
		Items Items `json:"items"`
	}

	var (
		resMergedMaster  Response
		resMergedRelease Response
		resOpenMaster    Response
		resOpenRelease   Response
	)

	since := releaseToTimestamp(app.config.base)
	until := releaseToTimestamp(app.config.head)

	url := fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:merged+closed:%s..%s+base:master", repo, since, until)
	if err := app.github.Get(url, &resMergedMaster); err != nil {
		fmt.Printf("error fetching issues merged into master branch of %s: %v", repo, err)
		os.Exit(1)
	}

	if app.config.head != "" {
		url = fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:merged+base:%s", repo, app.config.head)
		if err := app.github.Get(url, &resMergedRelease); err != nil {
			fmt.Printf("error fetching issues merged into release branch [%s] of %s: %v", app.config.head, repo, err)
			os.Exit(1)
		}
		url = fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:open+base:%s", repo, app.config.head)
		if err := app.github.Get(url, &resOpenRelease); err != nil {
			fmt.Printf("error fetching open issues for release branch [%s] of %s : %v", app.config.head, repo, err)
			os.Exit(1)
		}
	}

	url = fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:open+base:master", repo)
	if err := app.github.Get(url, &resOpenMaster); err != nil {
		fmt.Printf("error fetching open issues for master branch of %s: %v", repo, err)
		os.Exit(1)
	}

	for _, item := range resMergedMaster.Items {
		issues = append(issues, item.toIssue(repo, false, true))
	}
	for _, item := range resMergedRelease.Items {
		issues = append(issues, item.toIssue(repo, true, true))
	}
	for _, item := range resOpenMaster.Items {
		issues = append(issues, item.toIssue(repo, false, false))
	}
	for _, item := range resOpenRelease.Items {
		issues = append(issues, item.toIssue(repo, true, false))
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
	const notesRegexp = "(?:##+\\s*|\\*\\*)(?i:release\\s*notes)\\**[\\r\\n]+((?s:.)*?)(?:\\z|\\*\\*|##+|JIRA: \\[)"
	re := regexp.MustCompile(notesRegexp)
	matches := re.FindStringSubmatch(description)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(strings.TrimRight(strings.Replace(matches[1], "\r", "", -1), "\n\r "))
}

func getTesting(description string) string {
	const notesRegexp = "(?:##+\\s*|\\*\\*)(?i:testing)\\**[\\r\\n]+((?s:.)*?)(?:\\z|\\*\\*|##+|JIRA: \\[)"
	re := regexp.MustCompile(notesRegexp)
	matches := re.FindStringSubmatch(description)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(strings.TrimRight(strings.Replace(matches[1], "\r", "", -1), "\n\r "))
}
