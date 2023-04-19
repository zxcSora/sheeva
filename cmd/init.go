package cmd

import (
	"sheeva/config"

	gitlab "github.com/xanzy/go-gitlab"
)

var (
	baseURL, token, rootDir string
	gitlabClient            *gitlab.Client
	groups, projects        []config.GitlabElement
)

func init() {
	baseURL, token, rootDir = config.LoadConfig()
	c, err := config.CreateGitlabClient(token, baseURL)
	if err != nil {
		panic(err)
	}

	gitlabClient = c
	groups, projects, err = config.ParseYaml(rootDir)
	if err != nil {
		panic(err)
	}
}

// При 40 тоже апишка страдает, но все таки отвечает на "тестовом гитлабе"
// Elapsed time: 6:11.89
// func manageProjects(project []config.GitlabElement, client *gitlab.Client) {
// 	sem := semaphore.NewWeighted(40) // ограничение на 40 горутин одновременно
// 	var wg sync.WaitGroup
// 	wg.Add(len(project))

// 	for _, p := range project {
// 		sem.Acquire(context.Background(), 1) // захват семафора
// 		go func(p config.GitlabElement) {
// 			defer sem.Release(1) // освобождение семафора
// 			defer wg.Done()
// 			ManageProject(p, client)
// 		}(p)
// 	}

// 	wg.Wait()
// }

// Так скорее всего задудосим гитлаб))
// func manageProjects(project []config.GitlabElement, client *gitlab.Client) {
// 	var wg sync.WaitGroup
// 	for _, p := range project {
// 		wg.Add(1)
// 		go func(p config.GitlabElement, client *gitlab.Client) {
// 			defer wg.Done()
// 			cmd.ManageProject(p, client)
// 		}(p, client)
// 	}
// 	wg.Wait()
// }
