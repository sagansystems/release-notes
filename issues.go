package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const ITEMS_PER_PAGE = 100

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

	since := releaseToTimestamp(app.config.base)
	until := releaseToTimestamp(app.config.head)

	query := fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:merged+merged:%s..%s+base:master", repo, since, until)
	if addIssues, err := app.getIssuesForQuery(query, repo, false, true); err != nil {
		fmt.Printf("error fetching issues merged into master branch of %s: %v", repo, err)
		os.Exit(1)
	} else {
		issues = append(issues, addIssues...)
	}

	if app.config.head != "" {
		query = fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:merged+base:%s", repo, app.config.head)
		if addIssues, err := app.getIssuesForQuery(query, repo, true, true); err != nil {
			fmt.Printf("error fetching issues merged into release branch [%s] of %s: %v", app.config.head, repo, err)
			os.Exit(1)
		} else {
			issues = append(issues, addIssues...)
		}
		query = fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:open+base:%s", repo, app.config.head)
		if addIssues, err := app.getIssuesForQuery(query, repo, false, false); err != nil {
			fmt.Printf("error fetching open issues for release branch [%s] of %s : %v", app.config.head, repo, err)
			os.Exit(1)
		} else {
			issues = append(issues, addIssues...)
		}
	}

	query = fmt.Sprintf("search/issues?q=repo:sagansystems/%s+is:open+base:master", repo)
	if addIssues, err := app.getIssuesForQuery(query, repo, true, false); err != nil {
		fmt.Printf("error fetching open issues for master branch of %s: %v", repo, err)
		os.Exit(1)
	} else {
		issues = append(issues, addIssues...)
	}

	return issues
}

func (app IssuesApp) getIssuesForQuery(url, repo string, isHotfix, isMerged bool) ([]Issue, error) {
	type Response struct {
		TotalCount int   `json:"total_count"`
		Items      Items `json:"items"`
	}

	var issues []Issue

	for page := 1; true; page++ {
		var res Response
		path := fmt.Sprintf("%s&page=%d&per_page=%d", url, page, ITEMS_PER_PAGE)
		if err := app.github.Get(path, &res); err != nil {
			return nil, err
		}
		for _, item := range res.Items {
			issues = append(issues, item.toIssue(repo, isHotfix, isMerged))
		}
		if res.TotalCount-ITEMS_PER_PAGE*page <= 0 {
			break
		}
	}

	return issues, nil
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
