package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Github struct {
	username string
	token    string
}

func NewGithubClient(username, token string) Github {
	return Github{username: username, token: token}
}

func (client Github) Get(path string, v interface{}) error {
	resp, err := http.Get(fmt.Sprintf("https://%v:%v@api.github.com/%v",
		client.username,
		client.token,
		path))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}

	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

func oauth2Client(token string) github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return *github.NewClient(tc)
}
