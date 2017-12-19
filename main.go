package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	config := parseFlags()
	github := Github{username: config.username, password: config.password}
	app := NewApp(config, github)

	lines := []string{}
	for _, pr := range app.PRs() {
		lines = append(lines,
			fmt.Sprintf("%s - https://github.com/%s/%s/pull/%s - %s", pr.Author, config.owner, pr.Repo, pr.Num, pr.Title))
	}

	sort.Strings(lines)
	fmt.Println(strings.Join(lines, "\n"))
}

type Config struct {
	base     string
	head     string
	owner    string
	password string
	repos    string
	username string
}

func parseFlags() (config Config) {
	flag.StringVar(&config.base, "since", "", "base branch for comparison (ex: release-20171212014000Z)")
	flag.StringVar(&config.head, "until", "master", "head branch for comparison")
	flag.StringVar(&config.password, "password", os.Getenv("GITHUB_TOKEN"), "(REQUIRED) github access token $GITHUB_TOKEN")
	flag.StringVar(&config.username, "username", os.Getenv("GITHUB_USER"), "(REQUIRED) github username $GITHUB_USER")
	flag.StringVar(&config.owner, "owner", "sagansystems", "github repo owner")
	flag.StringVar(&config.repos, "repos", strings.Join(strings.Fields(`
			agent-desktop
			chat-sdk
			checkup
			edge-broker
			edge-router
			help
			remail
			scoring
			scoring
			supernova
			synthetic
		`), " "), "github repo names")
	flag.Parse()

	if config.username == "" || config.password == "" || config.base == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	return
}

type Github struct {
	username string
	password string
}

func (config Github) Get(path string, v interface{}) error {
	// log.Printf("GET %s", path)

	resp, err := http.Get(fmt.Sprintf("https://%v:%v@api.github.com/%v",
		config.username,
		config.password,
		path))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

type GithubGateway interface {
	Get(path string, v interface{}) error
}

type App struct {
	config Config
	github GithubGateway
}

type PR struct {
	Author string
	Num    string
	Title  string
	Repo   string
}

func NewApp(config Config, github GithubGateway) App {
	return App{config, github}
}

func (app App) PRs() []PR {
	prs := []PR{}

	for _, repo := range strings.Fields(app.config.repos) {
		prs = append(prs, app.RepoPRs(repo)...)
	}

	return prs
}

func (app App) RepoPRs(repo string) []PR {
	prs := []PR{}

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

				prs = append(prs, PR{
					Author: commit.Commit.Author.Name,
					Repo:   repo,
					Num:    result["num"],
					Title:  result["title"],
				})

				break
			}
		}
	}

	return prs
}
