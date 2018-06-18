package main

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/google/go-github/github"
)

func parseBranchTemplate(config Config, client github.Client) {
	fmt.Printf("%s", config.base)

	parseBranch(config.base)

	releaseBranches := listReleaseBranches(client, "sagansystems", "agent-desktop")
	fmt.Printf("Last branch: %s\n", releaseBranches[len(releaseBranches)-1])
	fmt.Printf("%s", releaseBranches)

}

func parseBranch(branch string) {
	branchReferenceRegex := regexp.MustCompile(".*latest-?(?P<index>\\d*).*")
	matches := branchReferenceRegex.FindStringSubmatch(branch)
	if matches != nil {
		fmt.Printf("matches(%d): %s\n", len(matches), matches)
	} else {
		fmt.Printf("no match")
	}
}

func listReleaseBranches(client github.Client, owner string, repo string) (branchNames []string) {

	releaseBranchRegex := regexp.MustCompile("^release-(?P<release_timestamp>\\d+)Z")

	opt := github.ListOptions{PerPage: 30}
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
				fmt.Printf("match: %s\n", matches)
				branchNames = append(branchNames, *branch.Name)
			}
		}
		if resp.NextPage == 0 {
			return
		}
		opt.Page = resp.NextPage
	}
}
