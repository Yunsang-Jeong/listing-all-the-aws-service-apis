package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type GithubConfig struct {
	Onwer         string
	RepoName      string
	LatestTagName string
	LatestTagSha  string
}

type GithubRepoTreeItem struct {
	Path string `json:"path"`
	Url  string `json:"url"`
	Sha  string `json:"sha"`
}

type GithubRepoBlob struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type GithubRepoTag struct {
	Name   string            `json:"name"`
	Commit map[string]string `json:"commit"`
}

func GetGithubRepoTrees(owner string, repo string, branch string, directroy string) ([]GithubRepoTreeItem, error) {
	path := fmt.Sprintf("repos/%s/%s/git/trees/%s:%s?recursive=1", owner, repo, branch, url.QueryEscape(directroy))
	data, err := githubAPIWithGetMethod(path)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get trees (%s)", path)
	}

	tree := struct {
		Tree []GithubRepoTreeItem `json:"tree"`
	}{}

	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}

	return tree.Tree, nil
}

func GetGithubRepoBlobs(owner string, repo string, version string, filename string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, version, filename)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.Wrapf(err, "[%s %s]  %s", resp.StatusCode, resp.Status, string(data[:]))
	}

	return data, nil
}

func GetGithubRepoLatestTag(owner string, repoName string) (*GithubRepoTag, error) {
	path := fmt.Sprintf("repos/%s/%s/tags", owner, repoName)
	data, err := githubAPIWithGetMethod(path)
	if err != nil {
		return nil, err
	}

	tags := []GithubRepoTag{}
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, err
	}

	return &tags[0], nil
}

func githubAPIWithGetMethod(path string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/%s", path)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", resp.Status, string(data[:]))
	}

	return data, nil
}
