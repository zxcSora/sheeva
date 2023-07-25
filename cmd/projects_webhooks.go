package cmd

import (
	"sheeva/config"

	logger "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

const defaultWebHookUrl = "https://gitlab2matrix.bingo-boom.ru/"

func EditProjectWebhooks(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	if project.HooksFile != "" {
		fileHooks, err := config.ParseHooksFile(project.HooksFile)
		if err != nil {
			logger.WithFields(logger.Fields{
				"Error":   err,
				"Project": project.Namespace + "/" + project.Name,
			}).Error("Error ocured while parsing webhooks file")
		}
		project.Hooks = append(project.Hooks, fileHooks.Hooks...)
	}
	err := cleanUnmanagedWebHooks(projectId, project, client)
	if err != nil {
		return err
	}
	if project.HooksFile == "" && project.Hooks == nil {
		defaultHook := defaultWebHook()
		project.Hooks = append(project.Hooks, *defaultHook)
	}
	for _, webhook := range project.Hooks {
		_, _, err := client.Projects.AddProjectHook(projectId, getWebHookOptions(webhook))
		if err != nil {
			return err
		}
		logger.WithFields(logger.Fields{
			"Project": project.Namespace + "/" + project.Name,
		}).Debug("Project Webhooks Successfully Managed")

	}
	return nil
}

func getWebHookOptions(webhook config.Hook) *gitlab.AddProjectHookOptions {
	WebhookOpts := &gitlab.AddProjectHookOptions{
		URL:                      gitlab.String(webhook.URL),
		PushEvents:               gitlab.Bool(webhook.PushEvents),
		ConfidentialIssuesEvents: gitlab.Bool(webhook.ConfidentialIssuesEvents),
		ConfidentialNoteEvents:   gitlab.Bool(webhook.ConfidentialNoteEvents),
		DeploymentEvents:         gitlab.Bool(webhook.DeploymentEvents),
		EnableSSLVerification:    gitlab.Bool(webhook.EnableSSLVerification),
		IssuesEvents:             gitlab.Bool(webhook.IssuesEvents),
		JobEvents:                gitlab.Bool(webhook.JobEvents),
		MergeRequestsEvents:      gitlab.Bool(webhook.MergeRequestsEvents),
		NoteEvents:               gitlab.Bool(webhook.NoteEvents),
		PipelineEvents:           gitlab.Bool(webhook.PipelineEvents),
		PushEventsBranchFilter:   gitlab.String(webhook.PushEventsBranchFilter),
		ReleasesEvents:           gitlab.Bool(webhook.ReleasesEvents),
		TagPushEvents:            gitlab.Bool(webhook.TagPushEvents),
		WikiPageEvents:           gitlab.Bool(webhook.WikiPageEvents),
		Token:                    gitlab.String(webhook.Token),
	}
	return WebhookOpts
}

func defaultWebHook() *config.Hook {
	return &config.Hook{
		URL:                      defaultWebHookUrl,
		PushEvents:               true,
		ConfidentialIssuesEvents: true,
		ConfidentialNoteEvents:   true,
		DeploymentEvents:         true,
		EnableSSLVerification:    true,
		IssuesEvents:             true,
		JobEvents:                true,
		MergeRequestsEvents:      true,
		NoteEvents:               true,
		PipelineEvents:           true,
		PushEventsBranchFilter:   "",
		ReleasesEvents:           true,
		TagPushEvents:            true,
		WikiPageEvents:           true,
		Token:                    "",
	}
}

func cleanUnmanagedWebHooks(projectId int, project config.GitlabElement, client *gitlab.Client) error {
	listProjectHooks, err := listProjectHooks(projectId, client)
	if err != nil {
		return err
	}
	for _, webhook := range listProjectHooks {
		projectHook, err := getProjectHook(projectId, webhook, client)
		if err != nil {
			return err
		} else {
			_, err = client.Projects.DeleteProjectHook(projectId, projectHook.ID)
		}
	}
	return nil
}

func getProjectHook(projectId int, ProjectHook *gitlab.ProjectHook, client *gitlab.Client) (*gitlab.ProjectHook, error) {
	projectHookData, _, err := client.Projects.GetProjectHook(projectId, ProjectHook.ID)
	if err != nil {
		return nil, err
	}

	return projectHookData, nil
}

func listProjectHooks(projectId int, client *gitlab.Client) ([]*gitlab.ProjectHook, error) {
	var listProjectHooks []*gitlab.ProjectHook
	listProjectHooks, _, err := client.Projects.ListProjectHooks(projectId, &gitlab.ListProjectHooksOptions{})
	if err != nil {
		return nil, err
	}
	return listProjectHooks, nil
}

func deleteProjectHook(projectId int, ProjectHook *gitlab.ProjectHook, client *gitlab.Client) error {
	var err error
	_, err = client.Projects.DeleteProjectHook(projectId, ProjectHook.ID)
	if err != nil {
		return err
	}
	return nil
}
