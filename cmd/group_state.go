package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func CreateGroup(parentID int, group config.GitlabElement, client *gitlab.Client) error {
	opts := createGroupOptions(group)

	switch group.Name {
	case group.Namespace:
		_, _, err := client.Groups.CreateGroup(opts)
		if err != nil {
			return err
		}
	default:
		opts.ParentID = &parentID
		_, _, err := client.Groups.CreateGroup(opts)
		if err != nil {
			return err
		}
	}

	// TODO: Надо переписать без циклического вызова
	manageGroup(group, client)
	return nil
}

func createGroupOptions(group config.GitlabElement) *gitlab.CreateGroupOptions {
	GroupOpts := &gitlab.CreateGroupOptions{
		Name:        gitlab.String(group.Name),
		Path:        gitlab.String(group.Name),
		Description: gitlab.String(group.Description),
	}
	return GroupOpts
}

func DeleteGroup(group config.GitlabElement, groupID int, client *gitlab.Client) error {
	if _, err := client.Groups.DeleteGroup(groupID); err != nil {
		logger.WithFields(logger.Fields{
			"Error":   err,
			"Group":   group.Name,
			"GroupID": groupID,
		}).Warning("Error while deleting namespace")
		return err
	}
	logger.WithField("Namespace", group.Name).Info("Group successfully deleted! ")
	return nil
}
