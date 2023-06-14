package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func ManageSchedules(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	scheduleList, err := listPipelineSchedules(projectId, client)
	if err != nil {
		logger.WithFields(logger.Fields{
			"Error":   err,
			"Project": project.Namespace + "/" + project.Name,
		}).Debug("Error ocured while getting project pipeline schedule list")
		return err
	}
	for _, schedule := range scheduleList {
		err := cleanUnmanagedPipelineSchedulest(projectId, schedule, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": project.Namespace + "/" + project.Name,
			}).Debug("Error ocured while remove unamanaged pipeline schedules")
			return err
		}
	}

	for _, sched := range project.Sched {
		Schedule, err := createPipelineSchedule(projectId, sched, client)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": project.Namespace + "/" + project.Name,
			}).Debug("Error ocured while creating project pipeline schedule")
			return err
		} else {
			for _, variable := range sched.Variables {
				err := createPipelineScheduleVariable(projectId, Schedule.ID, variable, client)
				if err != nil {
					logger.WithFields(logger.Fields{
						"Error":   err,
						"Project": project.Namespace + "/" + project.Name,
					}).Debug("Error ocured while creating project pipeline schedule variable")
					return err
				}
			}
		}
	}

	return nil
}

func listPipelineSchedulesOptions() *gitlab.ListPipelineSchedulesOptions {
	ListPipelineSchedulesOptions := &gitlab.ListPipelineSchedulesOptions{}
	return ListPipelineSchedulesOptions
}

func listPipelineSchedules(projectId int, client *gitlab.Client) ([]*gitlab.PipelineSchedule, error) {
	schedules, _, err := client.PipelineSchedules.ListPipelineSchedules(projectId, listPipelineSchedulesOptions())
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func cleanUnmanagedPipelineSchedulest(projectId int, schedule *gitlab.PipelineSchedule, client *gitlab.Client) error {
	_, err := client.PipelineSchedules.DeletePipelineSchedule(projectId, schedule.ID)
	if err != nil {
		return err
	}
	return nil
}

func createPipelineScheduleOptions(schedule config.Sched) *gitlab.CreatePipelineScheduleOptions {
	CreatePipelineSchedulesOptions := &gitlab.CreatePipelineScheduleOptions{
		Description: gitlab.String(schedule.Description),
		Cron:        gitlab.String(schedule.Cron),
		Ref:         gitlab.String(schedule.Ref),
	}
	return CreatePipelineSchedulesOptions
}

func createPipelineSchedule(projectId int, schedule config.Sched, client *gitlab.Client) (*gitlab.PipelineSchedule, error) {
	Schedule, _, err := client.PipelineSchedules.CreatePipelineSchedule(projectId, createPipelineScheduleOptions(schedule))
	if err != nil {
		return nil, err
	}
	return Schedule, nil
}

func createPipelineScheduleVariableOptions(variable config.Variable) *gitlab.CreatePipelineScheduleVariableOptions {
	CreatePipelineScheduleVariableOption := &gitlab.CreatePipelineScheduleVariableOptions{
		Key:          gitlab.String(variable.Key),
		Value:        gitlab.String(variable.Value),
		VariableType: gitlab.String(variable.VariableType),
	}
	return CreatePipelineScheduleVariableOption
}

func createPipelineScheduleVariable(projectId int, scheduleID int, variable config.Variable, client *gitlab.Client) error {
	_, _, err := client.PipelineSchedules.CreatePipelineScheduleVariable(projectId, scheduleID, createPipelineScheduleVariableOptions(variable))
	if err != nil {
		return err
	}
	return nil
}
