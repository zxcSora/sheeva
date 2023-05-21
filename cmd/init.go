package cmd

import (
	"sheeva/config"

	gitlab "github.com/xanzy/go-gitlab"
)

var (
	baseURL, token, rootDir string
	gitlabClient            *gitlab.Client
	groups, projects        []config.GitlabElement
)

func init() {
	baseURL, token, rootDir = config.LoadConfig()
	c, err := config.CreateGitlabClient(token, baseURL)
	if err != nil {
		panic(err)
	}

	gitlabClient = c
	groups, projects, err = config.ParseYaml(rootDir)
	if err != nil {
		panic(err)
	}
}
