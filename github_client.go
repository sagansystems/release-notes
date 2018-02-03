package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
