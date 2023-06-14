package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func manageFreezePeriods(group config.GitlabElement, client *gitlab.Client) error {
	if group.DeployFreezes != nil {
		var groupFullPath string
		if group.Name == group.Namespace {
			groupFullPath = group.Name
		} else {
			groupFullPath = group.Namespace + "/" + group.Name
		}
		CurrentGroupID, err := GetGroupID(groupFullPath, client)
		if err != nil {
			return err
		}
		ListRootGroupProjects, err := listGroupProjects(CurrentGroupID, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": groupFullPath,
			}).Error("Error while recieving project in group")
			return err
		}

		ListRootGroupSubGroups, err := listSubGroups(CurrentGroupID, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": groupFullPath,
			}).Error("Error while recieving subgroups in group")
			return err
		}
		for _, Subgroup := range ListRootGroupSubGroups {

			projects, err := listAllProjects(Subgroup.ID, client)

			if err != nil {
				logger.WithFields(logger.Fields{
					"Error": err,
					"Group": groupFullPath,
				}).Error("Error while recieving projects in subgroups")
				return err
			}
			ListRootGroupProjects = append(ListRootGroupProjects, projects...)
		}

		for _, project := range ListRootGroupProjects {
			for _, freezePeriod := range group.DeployFreezes {
				err := CreateFreezePeriod(project.ID, freezePeriod, client)
				if err != nil {
					logger.WithFields(logger.Fields{
						"Error": err,
						"Group": groupFullPath,
					}).Error("Error while creating freeze periods")
				}
			}
		}

		logger.WithFields(logger.Fields{
			"Group": groupFullPath,
			"State": group.State,
		}).Info("Freeze period successfully Managed")
	}
	return nil
}

func listAllProjects(groupID int, client *gitlab.Client) ([]*gitlab.Project, error) {
	projects, err := listGroupProjects(groupID, client)
	if err != nil {
		return nil, err
	}

	subGroups, err := listSubGroups(groupID, client)
	if err != nil {
		return nil, err
	}

	for _, subGroup := range subGroups {
		subGroupProjects, err := listAllProjects(subGroup.ID, client)
		if err != nil {
			return nil, err
		}
		projects = append(projects, subGroupProjects...)
	}

	return projects, nil
}

func listGroupProjects(groupID int, client *gitlab.Client) ([]*gitlab.Project, error) {
	ListProjects, _, err := client.Groups.ListGroupProjects(groupID, nil)
	if err != nil {
		return nil, err
	}
	return ListProjects, nil
}

func listSubGroups(groupID int, client *gitlab.Client) ([]*gitlab.Group, error) {
	ListSubGroups, _, err := client.Groups.ListSubGroups(groupID, nil)
	if err != nil {
		return nil, err
	}
	return ListSubGroups, nil
}

func CreateFreezePeriod(projectID int, freezePeriod config.DeployFreeze, client *gitlab.Client) error {
	_, _, err := client.FreezePeriods.CreateFreezePeriodOptions(projectID, CreateFreezePeriodOptions(freezePeriod))
	if err != nil {
		return err
	}
	return nil
}

func CreateFreezePeriodOptions(freezePeriod config.DeployFreeze) *gitlab.CreateFreezePeriodOptions {
	CreateFreezePeriodOptions := &gitlab.CreateFreezePeriodOptions{
		FreezeStart:  gitlab.String(freezePeriod.FreezeStart),
		FreezeEnd:    gitlab.String(freezePeriod.FreezeEnd),
		CronTimezone: gitlab.String(freezePeriod.CronTimezone),
	}
	return CreateFreezePeriodOptions
}
func ListFreezePeriods(projectID int, client *gitlab.Client) ([]*gitlab.FreezePeriod, error) {
	ListFreezePeriod, _, err := client.FreezePeriods.ListFreezePeriods(projectID, ListFreezePeriodsOptions())
	if err != nil {
		return nil, err
	}
	return ListFreezePeriod, nil
}

func ListFreezePeriodsOptions() *gitlab.ListFreezePeriodsOptions {
	ListFreezePeriodsOptions := &gitlab.ListFreezePeriodsOptions{}
	return ListFreezePeriodsOptions
}
func CleanUnmanagedFreezePeriods(projectID, freezePeriodID int, client *gitlab.Client) error {
	_, err := client.FreezePeriods.DeleteFreezePeriod(projectID, freezePeriodID)
	if err != nil {
		return err
	}
	return nil
}
