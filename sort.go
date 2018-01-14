package main

import "strings"

type IssuesByAuthor []Issue

func (iba IssuesByAuthor) Len() int {
	return len(iba)
}

func (iba IssuesByAuthor) Less(i, j int) bool {
	return strings.Compare(iba[i].Author, iba[j].Author) == -1
}

func (iba IssuesByAuthor) Swap(i, j int) {
	tmp := iba[i]
	iba[i] = iba[j]
	iba[j] = tmp
}
