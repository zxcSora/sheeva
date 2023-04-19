package main

import (
	"sheeva/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Recovered: ", r)
		}
	}()

	if err := cmd.ManageGroups(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Error while managing groups")
	}

	if err := cmd.ManageProjects(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Error while managing projects")
	}
}
