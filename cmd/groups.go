package cmd

import (
	"fmt"
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func manageGroup(group config.GitlabElement, client *gitlab.Client) error {
	var groupFullPath string
	if group.Name == group.Namespace {
		groupFullPath = group.Name
	} else {
		groupFullPath = group.Namespace + "/" + group.Name
	}

	parentID, err := GetGroupID(group.Namespace, client)
	if groupID, err := GetGroupID(groupFullPath, client); groupID == -1 && err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"Group": groupFullPath,
		}).Warning("Group Not Found")
		if group.State == "present" {
			if err := CreateGroup(parentID, group, client); err != nil {
				logger.WithFields(logger.Fields{
					"Error": err,
					"Group": groupFullPath,
				}).Error("Error while creating group")
				return err
			}
		}
	}
	groupID, err := GetGroupID(groupFullPath, client)

	if group.State == "absent" {
		return DeleteGroup(group, groupID, client)
	}

	if group.Avatar != "" {
		if err := UploadGroupAvatar(groupID, group.Avatar, client); err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": groupFullPath,
			}).Error("Error while updating group avatar")
		}
	}

	ManageVariables(groupID, group, client)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"Group": groupFullPath,
		}).Error("Error while recieving group data")
	}
	logger.WithFields(logger.Fields{
		"Group": groupFullPath,
		"State": group.State,
	}).Info("Group successfully Managed")

	return nil
}

func getGroupData(groupID int, client *gitlab.Client) (*gitlab.Group, error) {
	parent, _, err := client.Groups.GetGroup(groupID, nil)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":    err,
			"Group ID": groupID,
		}).Error(err)
		return nil, err
	}
	return parent, nil
}

type GroupNotFoundError struct {
	group string
}

func GroupNotFoundErrorWithGroup(v string) GroupNotFoundError {
	return GroupNotFoundError{v}
}

func (g GroupNotFoundError) Error() string {
	return fmt.Sprintf("Group '%s' not found", g.group)
}

func GetGroupID(groupPath string, client *gitlab.Client) (int, error) {
	group, _, err := client.Groups.GetGroup(groupPath, nil)
	if err != nil {
		return -1, GroupNotFoundErrorWithGroup(groupPath)
	}

	return group.ID, nil

}
