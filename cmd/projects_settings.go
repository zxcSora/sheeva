package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func EditProjectSetting(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	//https://pkg.go.dev/github.com/xanzy/go-gitlab#EditProjectOptions
	_, _, err := client.Projects.EditProject(projectId, editProjectSettingOpts(project))
	if err != nil {
		return err
	}

	logger.WithFields(logger.Fields{
		"Project": project.Namespace + "/" + project.Name,
	}).Debug("Project Settings Successfully Managed")
	return nil
}

func editProjectSettingOpts(project config.GitlabElement) *gitlab.EditProjectOptions {
	return &gitlab.EditProjectOptions{
		CIConfigPath:                 gitlab.String(project.CIConfigPath),
		RemoveSourceBranchAfterMerge: gitlab.Bool(false),
		CIForwardDeploymentEnabled:   gitlab.Bool(false),
	}
}
