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

func manageProject(project config.GitlabElement, client *gitlab.Client) error {
	if project.NamespaceOld != "" {
		TransferProject(project, client)
	}
	projectId, _ := GetProjectId(project.Namespace+"/"+project.Name, client)
	groupID, _ := GetGroupID(project.Namespace, client)

	switch project.State {
	case "present":
		switch projectId {
		case -1:
			_, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
				Name:                 gitlab.String(project.Name),
				Description:          gitlab.String(project.Description),
				Path:                 gitlab.String(project.Name),
				NamespaceID:          gitlab.Int(groupID),
				InitializeWithReadme: gitlab.Bool(true),
				DefaultBranch:        gitlab.String("master"),
			})
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":   err,
					"Project": project.Namespace + "/" + project.Name,
				}).Error("Error while creating project")
			}
		default:
			_, _, _ = client.Projects.UnarchiveProject(projectId, nil)
			logger.WithFields(logger.Fields{
				"Project": project.Namespace + "/" + project.Name,
			}).Debug("Project already exists")
		}
		pId, _ := GetProjectId(project.Namespace+"/"+project.Name, client)
		if project.Avatar != "" {
			UploadProjectAvatar(pId, project, client)
		}
		ManageProjectVariables(pId, project, client)
		EditProjectSetting(pId, project, client)
		ManageSchedules(pId, project, client)
		logger.WithFields(logger.Fields{
			"Project": project.Namespace + "/" + project.Name,
			"ID":      pId,
			"State":   project.State,
		}).Info("Project successfully Managed")
	case "archive":
		_, _, err := client.Projects.ArchiveProject(projectId, nil)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
			}).Warning("Error occured")
		} else {
			logger.WithField("Project", project.Namespace+"/"+project.Name).Warning("Project successfully achived! ")
		}
	case "absent":
		_, err := client.Projects.DeleteProject(projectId)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
			}).Warning("Error while deleting project")
		} else {
			logger.WithField("Project", project.Namespace+"/"+project.Name).Warning("Project successfully deleted! ")
		}
	}
	return nil
}

func ManageProjectVariables(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	if err := CleanUnmanagedVariablesProject(projectId, client); err != nil {
		logger.WithFields(logger.Fields{
			"Error":   err,
			"Project": project.Namespace + "/" + project.Name,
		}).Warning("Error ocured while remove unamanaged variables")
	}

	variablesFile := project.VariablesFile
	if variablesFile != "" {
		filePath := filepath.Join(variablesFile)
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			logger.Error(err)
			return err
		}
		var fileVariables struct {
			Variables []config.Variable `yaml:"variables,omitempty"`
		}
		err = yaml.Unmarshal(fileBytes, &fileVariables)
		if err != nil {
			logger.Error(err)
			return err
		}
		project.Variables = append(project.Variables, fileVariables.Variables...)
	}

	for _, variable := range project.Variables {
		_, _, err := client.ProjectVariables.CreateVariable(projectId, &gitlab.CreateProjectVariableOptions{
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
				"Project":  project.Namespace + "/" + project.Name,
				"Variable": variable.Key,
			}).Warning("Error ocured while create variable")
		} else {
			logger.WithFields(logger.Fields{
				"Project":       project.Namespace + "/" + project.Name,
				"Variable Name": variable.Key,
			}).Debug("Project Variable Successfully Managed")
		}
	}
	return nil
}
func GetProjectId(projectPath string, client *gitlab.Client) (int, error) {
	project, _, err := client.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{})
	if err != nil {
		return -1, err
	}
	return project.ID, nil
}

func CleanUnmanagedVariablesProject(projectId int, client *gitlab.Client) error {
	vars, _, err := client.ProjectVariables.ListVariables(projectId, &gitlab.ListProjectVariablesOptions{})
	if err != nil {
		return err
	}
	for _, v := range vars {
		_, err := client.ProjectVariables.RemoveVariable(projectId, v.Key, &gitlab.RemoveProjectVariableOptions{
			Filter: &gitlab.VariableFilter{EnvironmentScope: v.EnvironmentScope},
		})
		if err != nil {
			return err
		}

	}
	return nil
}

func UploadProjectAvatar(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	avatar, err := os.Open(project.Avatar)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
			"File":  project.Avatar,
		}).Error("Error occured")
		return nil
	}
	defer avatar.Close()

	_, _, err = client.Projects.UploadAvatar(projectId, avatar, project.Avatar, nil)

	if err != nil {
		logger.Error(err)
		return nil
	} else {
		logger.WithFields(logger.Fields{
			"Project": project.Namespace + "/" + project.Name,
		}).Debug("Project Avatar Successfully Managed")
	}
	return nil
}
func EditProjectSetting(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	//https://pkg.go.dev/github.com/xanzy/go-gitlab#EditProjectOptions
	_, _, err := client.Projects.EditProject(projectId, &gitlab.EditProjectOptions{
		CIConfigPath:                 gitlab.String(project.CIConfigPath),
		RemoveSourceBranchAfterMerge: gitlab.Bool(false),
		CIForwardDeploymentEnabled:   gitlab.Bool(false),
	})
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
		}).Warning("Error occured")
	} else {
		logger.WithFields(logger.Fields{
			"Project": project.Namespace + "/" + project.Name,
		}).Debug("Project Settings Successfully Managed")
	}
	return nil
}
func ManageSchedules(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	schedules, _, err := client.PipelineSchedules.ListPipelineSchedules(projectId, &gitlab.ListPipelineSchedulesOptions{})
	if err != nil {
		logger.Error(err)
	}
	for _, schedule := range schedules {
		_, err := client.PipelineSchedules.DeletePipelineSchedule(projectId, schedule.ID)
		if err != nil {
			logger.Error(err)
		}
	}
	if project.Sched != nil {
		for _, sched := range project.Sched {
			newSchedule, _, err := client.PipelineSchedules.CreatePipelineSchedule(projectId, &gitlab.CreatePipelineScheduleOptions{
				Description: &sched.Description,
				Cron:        &sched.Cron,
				Ref:         &sched.Ref,
			})
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":    err,
					"Schedule": sched.Description,
				}).Error("Error occured")
			} else {
				logger.WithFields(logger.Fields{
					"Schedule": sched.Description,
				}).Debug("Project Schedule Successfully Managed")
			}

			for _, variable := range sched.Variables {
				_, _, err := client.PipelineSchedules.CreatePipelineScheduleVariable(projectId, newSchedule.ID, &gitlab.CreatePipelineScheduleVariableOptions{
					Key:          &variable.Key,
					Value:        &variable.Value,
					VariableType: &variable.VariableType,
				})
				if err != nil {
					logger.WithFields(logger.Fields{
						"Error":    err,
						"Variable": variable.Key,
					}).Error("Error occured")
				} else {
					logger.WithFields(logger.Fields{
						"Schedule": sched.Description,
						"Variable": variable.Key,
					}).Debug("Project Schedule Variable Successfully Managed")
				}
			}
		}
	}
	return nil

}
func TransferProject(project config.GitlabElement, client *gitlab.Client) {
	projectId, _ := GetProjectId(project.NamespaceOld+"/"+project.Name, client)
	_, _, err := client.Projects.TransferProject(projectId, &gitlab.TransferProjectOptions{
		Namespace: &project.Namespace,
	})
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error": err,
		}).Warning("Error occured with transfer project")
	} else {
		logger.WithFields(logger.Fields{
			"Project": project.Namespace + "/" + project.Name,
			"ID":      projectId,
			"State":   project.State,
		}).Info("Project Successfully Transfered")
	}
}
