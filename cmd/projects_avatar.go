package cmd

import (
	"os"
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func UploadProjectAvatar(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	avatar, err := os.Open(project.Avatar)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"File":  project.Avatar,
		}).Debug("Error occured")
		return err
	}
	defer avatar.Close()

	_, _, err = client.Projects.UploadAvatar(projectId, avatar, project.Avatar, nil)
	if err != nil {
		logger.Warn(err)
		return err
	} else {
		logger.WithFields(logger.Fields{
			"Project": project.Namespace + "/" + project.Name,
		}).Debug("Project Avatar Successfully Managed")
	}
	return nil
}
