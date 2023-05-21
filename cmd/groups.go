package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	yaml "gopkg.in/yaml.v3"
)

func manageGroup(group config.GitlabElement, client *gitlab.Client) error {
	var groupFullPath string
	if group.Name == group.Namespace {
		groupFullPath = group.Name
	} else {
		groupFullPath = group.Namespace + "/" + group.Name
	}
	parentID, err := GetGroupID(group.Namespace, client)
	if groupID, err := GetGroupID(groupFullPath, client); groupID == -1 {
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": groupFullPath,
			}).Warning("Group Not Found")
			if group.State == "present" {
				if err := createGroup(parentID, group, client); err != nil {
					logger.WithFields(logger.Fields{
						"Error": err,
						"Group": groupFullPath,
					}).Error("Error while creating group")
					return err
				}
			}
		}
	}
	groupID, err := GetGroupID(groupFullPath, client)
	if group.State == "absent" {
		return deleteGroup(group, groupID, client)
	}
	if group.Avatar != "" {
		if err := uploadGroupAvatar(groupID, group.Avatar, client); err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": groupFullPath,
			}).Error("Error while updating group avatar")
		}
	}
	if group.Variables != nil || group.VariablesFile != "" {
		ManageVariables(groupID, group, client)
	}
	groupData, err := getGroupData(groupID, client)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"Group": groupFullPath,
		}).Error("Error while recieving group data")

	}
	if group.DeployFreezes != nil {
		manageFreezePeriod(client, groupData, group)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": groupFullPath,
			}).Error("Error while updating freeze period")

		}
	}
	logger.WithFields(logger.Fields{
		"Group": groupFullPath,
		"ID":    groupID,
		"State": group.State,
	}).Info("Group successfully Managed")

	return nil
}

func createGroup(parentID int, group config.GitlabElement, client *gitlab.Client) error {
	opts := gitlab.CreateGroupOptions{
		Name:        gitlab.String(group.Name),
		Path:        gitlab.String(group.Name),
		Description: gitlab.String(group.Description),
	}

	switch group.Name {
	case group.Namespace:
		_, _, err := client.Groups.CreateGroup(&opts)
		if err != nil {
			return err
		}
	default:
		opts.ParentID = &parentID
		_, _, err := client.Groups.CreateGroup(&opts)
		if err != nil {
			return err
		}
	}
	manageGroup(group, client)
	return nil
}

func deleteGroup(group config.GitlabElement, groupID int, client *gitlab.Client) error {
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

func ManageVariables(groupID int, group config.GitlabElement, client *gitlab.Client) error {
	if err := CleanUnmanagedVariablesGroup(groupID, client); err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"Group": group.Namespace + "/" + group.Name,
		}).Warning("Error ocured while remove unamanaged variables")
	}

	variablesFile := group.VariablesFile
	if variablesFile != "" {
		filePath := filepath.Join(variablesFile)
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"File":  variablesFile,
			}).Error("Error occured")
			return err
		}
		var fileVariables struct {
			Variables []config.Variable `yaml:"variables"`
		}
		err = yaml.Unmarshal(fileBytes, &fileVariables)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"File":  variablesFile,
			}).Error("Error occured")
			return err
		}
		group.Variables = append(group.Variables, fileVariables.Variables...)
	}

	for _, variable := range group.Variables {
		_, _, err := client.GroupVariables.CreateVariable(groupID, &gitlab.CreateGroupVariableOptions{
			Key:              gitlab.String(variable.Key),
			Value:            gitlab.String(variable.Value),
			VariableType:     gitlab.VariableType(gitlab.VariableTypeValue(variable.VariableType)),
			Protected:        gitlab.Bool(variable.Protected),
			Masked:           gitlab.Bool(variable.Masked),
			EnvironmentScope: gitlab.String(variable.Environment),
		})
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":    err,
				"Group":    group.Namespace + "/" + group.Name,
				"Variable": variable.Key,
			}).Warning("Error ocured while create variable")
		} else {
			logger.WithFields(logger.Fields{
				"Group":         group.Namespace + "/" + group.Name,
				"Variable Name": variable.Key,
			}).Debug("Group Variable Successfully Managed")
		}
	}
	return nil
}

func GetGroupID(groupPath string, client *gitlab.Client) (int, error) {
	group, _, err := client.Groups.GetGroup(groupPath, nil)
	if err != nil {
		return -1, err
	}

	return group.ID, nil

}

func CleanUnmanagedVariablesGroup(groupID int, client *gitlab.Client) error {
	vars, _, err := client.GroupVariables.ListVariables(groupID, &gitlab.ListGroupVariablesOptions{})
	if err != nil {
		return err
	}
	for _, v := range vars {
		_, err := client.GroupVariables.RemoveVariable(groupID, v.Key)
		if err != nil {
			return err
		}

	}
	return nil
}
func uploadGroupAvatar(groupId int, avatarFilePath string, client *gitlab.Client) error {
	avatar, err := os.Open(avatarFilePath)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"File":  avatarFilePath,
		}).Error("Error occured")
	}
	defer avatar.Close()

	_, _, err = client.Groups.UploadAvatar(groupId, avatar, avatarFilePath, nil)

	if err != nil {
		logger.Error(err)
	} else {
		logger.WithFields(logger.Fields{
			"Group": groupId,
		}).Debug("Group Avatar Successfully Managed")
	}

	return nil
}

func manageFreezePeriod(client *gitlab.Client, group *gitlab.Group, groups config.GitlabElement) {

	repos, _, err := client.Groups.ListGroupProjects(group.ID, nil)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":      err,
			"Group Name": group.Name,
		}).Error(err)
	}
	createFreeze(client, repos, groups)

	subgroups, _, err := client.Groups.ListSubGroups(group.ID, nil)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":      err,
			"Group Name": group.Name,
		}).Error(err)

	}

	for _, subgroup := range subgroups {
		manageFreezePeriod(client, subgroup, groups)
	}
}

func createFreeze(client *gitlab.Client, repos []*gitlab.Project, group config.GitlabElement) {
	for _, repo := range repos {
		freezePeriods, _, err := client.FreezePeriods.ListFreezePeriods(repo.ID, nil)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":      err,
				"Group Name": repo.Name,
			}).Error(err)
		}

		for _, fp := range freezePeriods {
			_, err := client.FreezePeriods.DeleteFreezePeriod(repo.ID, fp.ID)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Вагины`":    err,
					"Group Name": repo.Name,
				}).Error(err)
			}

		}
	}
	for _, repo := range repos {
		for _, freeze := range group.DeployFreezes {
			_, _, err := client.FreezePeriods.CreateFreezePeriodOptions(repo.ID, &gitlab.CreateFreezePeriodOptions{
				FreezeStart:  gitlab.String(freeze.FreezeStart),
				FreezeEnd:    gitlab.String(freeze.FreezeEnd),
				CronTimezone: gitlab.String(freeze.CronTimezone),
			})
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":      err,
					"Group Name": repo.Name,
				}).Error(err)
				continue
			}
		}
	}
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
