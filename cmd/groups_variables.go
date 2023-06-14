package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func ManageVariables(groupID int, group config.GitlabElement, client *gitlab.Client) {
	if group.CleanUnmanagedVars {
		if err := CleanUnmanagedVariablesGroup(groupID, client); err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Namespace + "/" + group.Name,
			}).Warning("Error ocured while remove unamanaged variables")
		}
	}

	variablesFile := group.VariablesFile
	if variablesFile != "" {
		fileVariables, err := config.ParseVariableFile(group.VariablesFile)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error": err,
				"Group": group.Namespace + "/" + group.Name,
			}).Error("Error ocured while parsing variable file")
		}
		group.Variables = append(group.Variables, fileVariables.Variables...)
	}

	for _, variable := range group.Variables {
		if err := CreateGroupVariable(groupID, variable, client); err != nil {
			err = UpdateGroupVariable(groupID, variable, client)
			if err != nil {
				logger.WithFields(logger.Fields{
					"Error":    err,
					"Group":    group.Namespace + "/" + group.Name,
					"Variable": variable.Key,
				}).Warning("Error ocured while updating variable")
			}
		}
	}
}

func CreateGroupVariable(groupID int, variable config.Variable, client *gitlab.Client) error {
	_, _, err := client.GroupVariables.CreateVariable(groupID, getCreateGroupVariableOptions(variable))
	if err != nil {
		return err
	}
	return nil
}

func getCreateGroupVariableOptions(variable config.Variable) *gitlab.CreateGroupVariableOptions {
	GroupVariableOpts := &gitlab.CreateGroupVariableOptions{
		Key:              gitlab.String(variable.Key),
		Value:            gitlab.String(variable.Value),
		VariableType:     gitlab.VariableType(gitlab.VariableTypeValue(variable.VariableType)),
		Protected:        gitlab.Bool(variable.Protected),
		Masked:           gitlab.Bool(variable.Masked),
		EnvironmentScope: gitlab.String(variable.Environment),
	}
	return GroupVariableOpts
}

func UpdateGroupVariable(groupID int, variable config.Variable, client *gitlab.Client) error {
	_, _, err := client.GroupVariables.UpdateVariable(groupID, variable.Key, getUpdateGroupVariableOptions(variable))
	if err != nil {
		return err
	}
	return nil
}

func getUpdateGroupVariableOptions(variable config.Variable) *gitlab.UpdateGroupVariableOptions {
	GroupVariableOpts := &gitlab.UpdateGroupVariableOptions{
		Value:            gitlab.String(variable.Value),
		VariableType:     gitlab.VariableType(gitlab.VariableTypeValue(variable.VariableType)),
		Protected:        gitlab.Bool(variable.Protected),
		Masked:           gitlab.Bool(variable.Masked),
		EnvironmentScope: gitlab.String(variable.Environment),
	}
	return GroupVariableOpts
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
