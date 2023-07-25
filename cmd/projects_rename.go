package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func renameProjectOptions(project config.GitlabElement) *gitlab.EditProjectOptions {
	return &gitlab.EditProjectOptions{
		Name: gitlab.String(project.Name),
		Path: gitlab.String(project.Name),
	}
}

func RenameProject(project config.GitlabElement, client *gitlab.Client) error {
	if project.Name == project.NameOld {
		return nil
	}

	projectId, _ := GetProjectId(project.Namespace+"/"+project.NameOld, client)
	if projectId < 0 {
		return nil
	}
	_, _, err := client.Projects.EditProject(projectId, renameProjectOptions(project))
	if err != nil {
		return err
	}

	logger.WithFields(logger.Fields{
		"Project": project.Namespace + "/" + project.Name,
		"NameOld": project.NameOld,
		"ID":      projectId,
		"State":   project.State,
	}).Debug("Project Successfully Renamed")
	return nil
}
