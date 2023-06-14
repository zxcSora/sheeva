package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func UnarchiveProject(projectId int, project config.GitlabElement, client *gitlab.Client) {
	_, _, _ = client.Projects.UnarchiveProject(projectId, nil)
	logger.WithFields(logger.Fields{
		"Project": project.Namespace + "/" + project.Name,
	}).Debug("Project already exists")
}

func ArchiveProject(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	_, _, err := client.Projects.ArchiveProject(projectId, nil)
	if err != nil {
		return err
	}
	logger.WithField("Project", project.Namespace+"/"+project.Name).Debug("Project successfully achived! ")
	return nil
}

func DeleteProject(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	_, err := client.Projects.DeleteProject(projectId)
	if err != nil {
		return err
	}
	logger.WithField("Project", project.Namespace+"/"+project.Name).Warning("Project successfully deleted! ")
	return nil
}

func CreateProject(groupID int, project config.GitlabElement, client *gitlab.Client) error {
	_, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:                 gitlab.String(project.Name),
		Description:          gitlab.String(project.Description),
		Path:                 gitlab.String(project.Name),
		NamespaceID:          gitlab.Int(groupID),
		InitializeWithReadme: gitlab.Bool(true),
		DefaultBranch:        gitlab.String(defaultBranch),
	})
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":   err,
			"Project": project.Namespace + "/" + project.Name,
		}).Error("Error while creating project")
	}
	return nil
}
