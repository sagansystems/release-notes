package main

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	base     string
	head     string
	owner    string
	password string
	repos    string
	username string

	hotfixOnly       bool
	useCommits       bool
	withInternal     bool
	withReleaseNotes bool
	withTesting      bool
}

func parseFlags() (config Config) {
	flag.StringVar(&config.base, "since", "", "base branch for comparison (specific like: release-20171212014000Z, or relative: latest-1)")
	flag.StringVar(&config.head, "until", "", "head branch for comparison (again, specific or relative like: latest)")
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
			supernova
			synthetic
			burrow
			elasticsearch-backup
			connector-service
			looker
			help-center
		`), " "), "github repo names")
	flag.BoolVar(&config.useCommits, "commits", false, "use commits instead of issues")
	flag.BoolVar(&config.withTesting, "with-testing", false, "include testing sections")
	flag.BoolVar(&config.withReleaseNotes, "with-release-notes", true, "include release notes sections")
	flag.BoolVar(&config.withInternal, "with-internal", true, "include items labeled as 'not release noted'")
	flag.BoolVar(&config.hotfixOnly, "hotfix-only", false, "show only hotfix issues")
	flag.Parse()

	if config.username == "" || config.password == "" || config.base == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	return
}
