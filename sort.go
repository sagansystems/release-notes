package main

import "strings"

type IssuesByAuthorAndInternal []Issue

func (iba IssuesByAuthorAndInternal) Len() int {
	return len(iba)
}

func (iba IssuesByAuthorAndInternal) Less(i, j int) bool {
	if contains(iba[i].Labels, notReleaseNoted) && !contains(iba[j].Labels, notReleaseNoted) {
		return false
	}
	if !contains(iba[i].Labels, notReleaseNoted) && contains(iba[j].Labels, notReleaseNoted) {
		return true
	}
	return strings.Compare(iba[i].Author, iba[j].Author) == -1
}

func (iba IssuesByAuthorAndInternal) Swap(i, j int) {
	tmp := iba[i]
	iba[i] = iba[j]
	iba[j] = tmp
}
