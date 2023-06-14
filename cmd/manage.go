package cmd

import (
	"os"
	"runtime"
	"sheeva/config"
	"sync"

	"strconv"

	log "github.com/sirupsen/logrus"
)

func maxGorutines() int {
	max := runtime.NumCPU() * 2

	if g := os.Getenv("SHEEVA_MAX_GORUTINES"); g != "" {
		i, err := strconv.Atoi(g)
		if err != nil {
			return max
		}
		max = i
	}

	return max
}

func ManageGroups() error {
	gg := NewGroupGraphs(groups)

	for _, g := range gg {
		// Create root groups
		if err := manageGroup(*g.Group, gitlabClient); err != nil {
			return err
		}

		var counter int
		for i := 1; ; i++ {
			nodes := g.GetEdgeNodes(i)
			if nodes == nil {
				break
			}

			wg := &sync.WaitGroup{}
			for _, node := range nodes {
				wg.Add(1)
				counter++
				go func(wg *sync.WaitGroup, node *Node) {
					defer wg.Done()
					if err := manageGroup(*node.Group, gitlabClient); err != nil {
						log.Fatal(err)
					}
				}(wg, node)

				if counter >= maxGorutines() {
					wg.Wait()
					counter = 0
				}
			}
			wg.Wait()
		}
	}

	return nil
}

func ManageProjects() error {
	wg := &sync.WaitGroup{}
	var counter int
	for _, p := range projects {
		wg.Add(1)
		counter++
		go func(wg *sync.WaitGroup, p config.GitlabElement) {
			defer wg.Done()
			if err := manageProject(p, gitlabClient); err != nil {
				log.Fatal(err)
			}
		}(wg, p)
		if counter >= maxGorutines() {
			wg.Wait()
			counter = 0
		}
	}
	wg.Wait()
	return nil
}

func ManageFreezePeriods() error {
	wg := &sync.WaitGroup{}
	var counter int
	for _, g := range groups {
		wg.Add(1)
		counter++
		go func(wg *sync.WaitGroup, g config.GitlabElement) {
			defer wg.Done()
			if err := manageFreezePeriods(g, gitlabClient); err != nil {
				log.Fatal(err)
			}
		}(wg, g)
		if counter >= maxGorutines() {
			wg.Wait()
			counter = 0
		}
	}
	wg.Wait()
	return nil
}
