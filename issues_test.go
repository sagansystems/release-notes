package main

import "testing"

func TestGetReleaseNotes(t *testing.T) {
	tables := []struct {
		input  string
		output string
	}{
		{"Release notes for this issue\nare not written yet", ""},
		{"## Release Notes\nmert mert", "mert mert"},
		{"**Release notes**\nalso work with stars", "also work with stars"},
		{"## RELEASE NOTES\ncase should not matter", "case should not matter"},
		{"**Release Notes**\nthere's also testing section that can follow\n**Testing", "there's also testing section that can follow"},
		{"## Release Notes \nsome people leave spaces after their section indicator\n# Test", "some people leave spaces after their section indicator"},
		{"# Release notes \nwe do support H1 too", "we do support H1 too"},
	}

	for _, table := range tables {
		releaseNotes := getReleaseNotes(table.input)
		if releaseNotes != table.output {
			t.Errorf("Release notes from:\n\"%s\"\nwere incorrect,\ngot:\n\t\"%s\",\nwant:\n\t\"%s.\"", table.input, releaseNotes, table.output)
		}
	}
}
