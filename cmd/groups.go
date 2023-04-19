package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	yaml "gopkg.in/yaml.v3"
)

func createGroup(parentId int, group config.GitlabElement, client *gitlab.Client) error {
	var gID int
	opts := gitlab.CreateGroupOptions{
		Name:        gitlab.String(group.Name),
		Path:        gitlab.String(group.Name),
		Description: gitlab.String(group.Description),
	}

	switch group.Name {
	case group.Namespace:
		groupId, err := GetGroupId(group.Name, client)
		if err != nil {
			_, _, err := client.Groups.CreateGroup(&opts)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":     err,
					"Namespace": group.Name,
				}).Error("Error while creating namespace")
			}
			groupId, err = GetGroupId(group.Name, client)
			if err != nil {
				return fmt.Errorf("Can't create a group: %s", err)
			}
		}

		// Мы действительно хотим видеть в логе всё про существующие
		// неймспейсы? Возможно, что стоит понизить уровень логирования до Debug
		logger.WithFields(logger.Fields{
			"Namespace": group.Name,
		}).Debug("Namespace already exists")
		gID = groupId

	default:
		opts.ParentID = &parentId
		groupId, err := GetGroupId(group.Namespace+"/"+group.Name, client)
		if err != nil {
			_, _, err := client.Groups.CreateGroup(&opts)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error": err,
					"Group": group.Namespace + "/" + group.Name,
				}).Error("Error while creating group")
			}
		}
		groupId, err = GetGroupId(group.Namespace+"/"+group.Name, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Namespace + "/" + group.Name,
			}).Error("Error while getting group ID")
		}
		gID = groupId

	}
	if group.Avatar != "" {
		go func(gID int, group config.GitlabElement, client *gitlab.Client) {
			if err := uploadGroupAvatar(gID, group.Avatar, client); err != nil {
				logger.WithFields(logger.Fields{
					"Error": err,
					"Group": group.Namespace + "/" + group.Name,
				}).Error("Error while updating group avatar")
			}
		}(gID, group, client)
	}
	if group.Variables != nil && group.VariablesFile != "" {
		go func(gID int, group config.GitlabElement, client *gitlab.Client) {
			ManageVariables(gID, group, client)
		}(gID, group, client)
	}
	groupData, err := getGroupData(gID, client)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"Group": group.Namespace + "/" + group.Name,
		}).Error("Error while updating freeze period")

	}
	manageFreezePeriod(client, groupData, group)
	logger.WithFields(logger.Fields{
		"Group": group.Namespace + "/" + group.Name,
		"ID":    gID,
		"State": group.State,
	}).Info("Group successfully Managed")
	return nil
}

func deleteGroup(group config.GitlabElement, client *gitlab.Client) error {
	switch group.Name {
	case group.Namespace:
		groupId, err := GetGroupId(group.Name, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Name,
			}).Warning("Namespace does not exist")
			return err
		}
		if _, err := client.Groups.DeleteGroup(groupId); err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Name,
			}).Warning("Error while deleting namespace")
			return err
		}
		logger.WithField("Namespace", group.Name).Info("Group successfully deleted! ")
	default:
		groupId, err := GetGroupId(group.Namespace+"/"+group.Name, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Name,
			}).Warning("Namespace does not exist")
			return err
		}
		if _, err := client.Groups.DeleteGroup(groupId); err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Namespace + "/" + group.Name,
			}).Warning("Error while deleting group")
			return err
		}
		logger.WithField("Group", group.Namespace+"/"+group.Name).Warning("Group successfully deleted! ")
	}
	return nil
}

func manageGroup(group config.GitlabElement, client *gitlab.Client) error {
	parentGroupId, _ := GetGroupId(group.Namespace, client)
	switch group.State {
	case "present":
		if err := createGroup(parentGroupId, group, client); err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Name,
			}).Warning("Error while creating namespace")
			return err
		}
	case "absent":
		return deleteGroup(group, client)
	default:
		return fmt.Errorf("Wrong state: %s", group.State)
	}
	return nil
}

func ManageVariables(groupID int, group config.GitlabElement, client *gitlab.Client) error {
	if err := CleanUnmanagedVariablesGroup(groupID, client); err != nil {
		return err
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
			logger.Error(err)
			return err
		} else {
			logger.WithFields(logger.Fields{
				"Group":         group.Namespace + "/" + group.Name,
				"Variable Name": variable.Key,
			}).Debug("Group Variable Successfully Managed")
		}
	}
	return nil
}

func GetGroupId(groupPath string, client *gitlab.Client) (int, error) {
	group, r, err := client.Groups.GetGroup(groupPath, nil)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":  err,
			"Status": r.Status,
		}).Error(err)
		return -1, err
	}

	return group.ID, nil

}

func CleanUnmanagedVariablesGroup(groupID int, client *gitlab.Client) error {
	vars, _, err := client.GroupVariables.ListVariables(groupID, &gitlab.ListGroupVariablesOptions{})
	if err != nil {
		logger.Error(err)
		return err
	}
	for _, v := range vars {
		_, err := client.GroupVariables.RemoveVariable(groupID, v.Key)
		if err != nil {
			logger.Error(v.Key, err)
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
		return err
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
		return
	}
	go createFreeze(client, repos, groups)

	subgroups, _, err := client.Groups.ListSubGroups(group.ID, nil)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":      err,
			"Group Name": group.Name,
		}).Error(err)
		return
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
			continue
		}
		for _, fp := range freezePeriods {
			_, err := client.FreezePeriods.DeleteFreezePeriod(repo.ID, fp.ID)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":      err,
					"Group Name": repo.Name,
				}).Error(err)
				continue
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
func getGroupData(gID int, client *gitlab.Client) (*gitlab.Group, error) {
	parent, _, err := client.Groups.GetGroup(gID, nil)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":    err,
			"Group ID": gID,
		}).Error(err)
		return nil, err
	}
	return parent, nil
}
