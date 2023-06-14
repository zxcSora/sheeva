package cmd

import (
	"fmt"
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

const defaultBranch = "master"

func manageProject(project config.GitlabElement, client *gitlab.Client) error {
	if project.NamespaceOld != "" {
		err := TransferProject(project, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
			}).Error("Error while transfering project")
		}
	}

	// TODO: Переписать с нормальной обработкой ошибок
	projectPath := project.Namespace + "/" + project.Name
	projectId, prErr := GetProjectId(projectPath, client)
	if prErr != nil {
		logger.Warnf("Project %s not found", projectPath)
	}

	groupID, _ := GetGroupID(project.Namespace, client)
	switch project.State {
	case "present":
		switch projectId {
		case -1:
			logger.Debugf("Project %s not found", projectPath)
			CreateProject(groupID, project, client)
		default:
			UnarchiveProject(projectId, project, client)
		}

		pId, _ := GetProjectId(project.Namespace+"/"+project.Name, client)
		if project.Avatar != "" {
			err := UploadProjectAvatar(pId, project, client)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":   err,
					"Project": projectPath,
				}).Error("Error while upload project avatar")
			}
		}
		ManageProjectVariables(pId, project, client)
		err := EditProjectSetting(pId, project, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": projectPath,
			}).Error("Error while edit project settings")
		}
		if project.Sched != nil {
			ManageSchedules(pId, project, client)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":   err,
					"Project": projectPath,
				}).Error("Error while managing project schedules")
			}
		}
		err = EditProjectWebhooks(pId, project, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Project": projectPath,
				"State":   project.State,
			}).Error("Error while managing project web hooks")
		}
		ListFreezePeriods, err := ListFreezePeriods(pId, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Project": projectPath,
				"State":   project.State,
			}).Error("Error while receive project freeze periods")
		}
		for _, freezePeriod := range ListFreezePeriods {
			err := CleanUnmanagedFreezePeriods(pId, freezePeriod.ID, client)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Project": projectPath,
					"State":   project.State,
				}).Error("Error while clean unamanged freeze periods")
			}

		}
	case "archive":
		err := ArchiveProject(projectId, project, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": projectPath,
			}).Error("Error while archive project")
		}
	case "absent":
		err := DeleteProject(projectId, project, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": projectPath,
			}).Error("Error while delete project")
		}
	}
	logger.WithFields(logger.Fields{
		"Project": projectPath,
		"State":   project.State,
	}).Info("Project successfully Managed")
	return nil
}

type ProjectNotFoundError struct {
	project string
}

func ProjectNotFoundErrorWithProject(v string) *ProjectNotFoundError {
	return &ProjectNotFoundError{v}
}

func (p *ProjectNotFoundError) Error() string {
	return fmt.Sprintf("Project '%s' not found", p.project)
}

func GetProjectId(projectPath string, client *gitlab.Client) (int, error) {
	project, _, err := client.Projects.GetProject(projectPath, getProjectOptions())
	if err != nil {
		return -1, ProjectNotFoundErrorWithProject(projectPath)
	}
	return project.ID, nil
}

func getProjectOptions() *gitlab.GetProjectOptions {
	GetProjectOptions := &gitlab.GetProjectOptions{}
	return GetProjectOptions
}
