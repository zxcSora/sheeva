package cmd

func ManageGroups() error {
	for _, g := range groups {
		if err := manageGroup(g, gitlabClient); err != nil {
			return err
		}
	}
	return nil
}

func ManageProjects() error {
	for _, p := range projects {
		if err := manageProject(p, gitlabClient); err != nil {
			return err
		}
	}
	return nil
}
