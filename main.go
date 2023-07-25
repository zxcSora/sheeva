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

func reportCaller() bool {
	if v := os.Getenv("SHEEVA_REPORT_CALLER"); v != "" {
		return v == "1"
	}
	return true
}

func disableColors() bool {
	return os.Getenv("SHEEVA_DISABLE_COLORS") == "1"
}

func debugEnable() bool {
	return os.Getenv("SHEEVA_DEBUG_ENABLE") == "1"
}

func init() {
	log.SetReportCaller(reportCaller())
	log.SetFormatter(&log.TextFormatter{
		ForceColors: !disableColors(),
	})

	if debugEnable() {
		log.SetLevel(log.DebugLevel)
	}
}
