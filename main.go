package main

import (
	"os"
	"sheeva/cmd"

	log "github.com/sirupsen/logrus"
)

func isRecoverDisabled() bool {
	return os.Getenv("SHEEVA_RECOVER_DISABLED") == "1"
}

func main() {
	defer func() {
		if !isRecoverDisabled() {
			if r := recover(); r != nil {
				log.Fatal("Recovered: ", r)
			}
		}
	}()

	if err := cmd.ManageGroups(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Error while managing groups")
	}

	if err := cmd.ManageProjects(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Error while managing projects")
	}
	if err := cmd.ManageFreezePeriods(); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Error while managing projects")
	}
}

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
}
