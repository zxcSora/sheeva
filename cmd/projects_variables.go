package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

// Можно параллелить вполне целиком эту функцию
func ManageProjectVariables(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	if project.CleanUnmanagedVars {
		if err := CleanUnmanagedVariablesProject(projectId, client); err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": project.Namespace + "/" + project.Name,
			}).Warning("Error ocured while remove unamanaged variables")
		}
	}

	variablesFile := project.VariablesFile
	if variablesFile != "" {
		fileVariables, err := config.ParseVariableFile(project.VariablesFile)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": project.Namespace + "/" + project.Name,
			}).Error("Error ocured while parsing variable file")
		}
		project.Variables = append(project.Variables, fileVariables.Variables...)
	}

	for _, variable := range project.Variables {
		err := CreateProjectVariable(projectId, variable, client)
		if err != nil {
			err = UpdateProjectVariable(projectId, variable, client)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":    err,
					"Project":  project.Namespace + "/" + project.Name,
					"Variable": variable.Key,
				}).Warning("Error ocured while create variable")
			}
		}
	}
	return nil
}

func CreateProjectVariable(projectID int, variable config.Variable, client *gitlab.Client) error {
	_, _, err := client.ProjectVariables.CreateVariable(projectID, createProjectVariableOptions(variable))
	if err != nil {
		return err
	}
	return nil
}

func createProjectVariableOptions(variable config.Variable) *gitlab.CreateProjectVariableOptions {
	ProjectVariableOpts := &gitlab.CreateProjectVariableOptions{
		Key:              gitlab.String(variable.Key),
		Value:            gitlab.String(variable.Value),
		VariableType:     gitlab.VariableType(gitlab.VariableTypeValue(variable.VariableType)),
		Protected:        gitlab.Bool(variable.Protected),
		Masked:           gitlab.Bool(variable.Masked),
		EnvironmentScope: gitlab.String(variable.Environment),
	}
	return ProjectVariableOpts
}

func UpdateProjectVariable(projectID int, variable config.Variable, client *gitlab.Client) error {
	_, _, err := client.ProjectVariables.UpdateVariable(projectID, variable.Key, updateProjectVariableOptions(variable))
	if err != nil {
		return err
	}
	return nil
}

func updateProjectVariableOptions(variable config.Variable) *gitlab.UpdateProjectVariableOptions {
	UpdateProjectVariableOpts := &gitlab.UpdateProjectVariableOptions{
		Value:            gitlab.String(variable.Value),
		VariableType:     gitlab.VariableType(gitlab.VariableTypeValue(variable.VariableType)),
		Protected:        gitlab.Bool(variable.Protected),
		Masked:           gitlab.Bool(variable.Masked),
		EnvironmentScope: gitlab.String(variable.Environment),
		Filter:           &gitlab.VariableFilter{EnvironmentScope: variable.Environment},
	}
	return UpdateProjectVariableOpts
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
