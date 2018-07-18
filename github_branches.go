package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"

	"github.com/google/go-github/github"
)

const PER_PAGE = 50

func parseBranchTemplate(config *Config, client github.Client) {

	// fetch branches, so that we don't have to repeat later
	releaseBranches := listReleaseBranches(client, "sagansystems", "agent-desktop")

	config.base = parseBranch(config.base, releaseBranches)

	config.head = parseBranch(config.head, releaseBranches)
}

func parseBranch(branch string, releaseBranches []string) string {
	var index = 0

	branchReferenceRegex := regexp.MustCompile(".*latest-?(?P<index>\\d*).*")
	matches := branchReferenceRegex.FindStringSubmatch(branch)
	if matches != nil {
		if len(matches[1]) > 0 {
			i, err := strconv.Atoi(matches[1])
			index = i
			if err != nil {
				fmt.Printf("error extracting index from %s: %v", matches[1], err)
				os.Exit(1)
			}
		}
		return releaseBranches[len(releaseBranches)-1-index]
	}

	// "latest..." not used, so just return what was passed
	return branch
}

func listReleaseBranches(client github.Client, owner string, repo string) (releaseBranches []string) {

	releaseBranchRegex := regexp.MustCompile("^release-(?P<release_timestamp>\\d+)Z")

	opt := github.ListOptions{PerPage: PER_PAGE}
	for {
		var (
			branches []*github.Branch
			resp     *github.Response
		)

		branches, resp, err := client.Repositories.ListBranches(context.Background(), owner, repo, &opt)
		if err != nil {
			fmt.Printf("error fetching release branches from %s: %v", repo, err)
			os.Exit(1)
		}
		for _, branch := range branches {
			matches := releaseBranchRegex.FindStringSubmatch(*branch.Name)
			if matches != nil {
				releaseBranches = append(releaseBranches, *branch.Name)
			}
		}
		if resp.NextPage == 0 {
			sort.Strings(releaseBranches) // make sure names are sorted
			return
		}
		opt.Page = resp.NextPage
	}
}
