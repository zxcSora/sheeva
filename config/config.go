package config

import (
	"flag"
	"os"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func LoadConfig() (string, string, string) {

	gitlabToken := os.Getenv("GITLAB_TOKEN")
	if gitlabToken == "" {
		flag.StringVar(&gitlabToken, "t", "", "Gitlab token not defined")
		flag.StringVar(&gitlabToken, "token", "", "Gitlab token not defined")

	}

	gitlabUrl := os.Getenv("GITLAB_URL")
	if gitlabUrl == "" {
		flag.StringVar(&gitlabUrl, "u", "", "Gitlab url not defined")
		flag.StringVar(&gitlabUrl, "url", "", "Gitlab url not defined")
	}

	rootDir := os.Getenv("ROOT_DIR")
	if gitlabUrl == "" {
		flag.StringVar(&rootDir, "d", "", "Root dir not defined")
		flag.StringVar(&rootDir, "dir", "", "Root dir not defined")
	}

	flag.Parse()
	if rootDir == "" {
		rootDir = "./projects"
	}
	return gitlabUrl, gitlabToken, rootDir
}
func CreateGitlabClient(gitlabToken, gitlabEndpoint string) (*gitlab.Client, error) {
	gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabEndpoint+"/api/v4"))
	if err != nil {
		logger.Warning(err)
		return nil, err
	}

	return gitlabClient, nil
}
