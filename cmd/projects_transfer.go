package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func transferProjectOptions(project config.GitlabElement) *gitlab.TransferProjectOptions {
	return &gitlab.TransferProjectOptions{
		Namespace: &project.Namespace,
	}
}

func TransferProject(project config.GitlabElement, client *gitlab.Client) error {
	projectId, _ := GetProjectId(project.NamespaceOld+"/"+project.Name, client)
	_, _, err := client.Projects.TransferProject(projectId, transferProjectOptions(project))
	if err != nil {
		return err
	}

	logger.WithFields(logger.Fields{
		"Project": project.Namespace + "/" + project.Name,
		"ID":      projectId,
		"State":   project.State,
	}).Debug("Project Successfully Transfered")
	return nil
}
