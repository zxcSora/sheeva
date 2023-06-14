package cmd

import (
	"os"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func UploadGroupAvatar(groupId int, avatarFilePath string, client *gitlab.Client) error {
	avatar, err := os.Open(avatarFilePath)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"File":  avatarFilePath,
		}).Error("Error occured")
		return err
	}
	defer avatar.Close()

	_, _, err = client.Groups.UploadAvatar(groupId, avatar, avatarFilePath, nil)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.WithFields(logger.Fields{
		"Group": groupId,
	}).Debug("Group Avatar Successfully Managed")

	return nil
}
