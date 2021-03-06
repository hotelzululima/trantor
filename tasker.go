package main

import (
	log "github.com/cihub/seelog"

	"time"

	"git.gitorious.org/trantor/trantor.git/database"
)

func InitTasks(db *database.DB) {
	periodicTask(updateLogger, MINUTES_UPDATE_LOGGER*time.Minute)
	periodicTask(db.UpdateTags, MINUTES_UPDATE_TAGS*time.Minute)
	periodicTask(db.UpdateMostVisited, MINUTES_UPDATE_VISITED*time.Minute)
	periodicTask(db.UpdateDownloadedBooks, MINUTES_UPDATE_DOWNLOADED*time.Minute)
	periodicTask(db.UpdateHourVisits, MINUTES_UPDATE_HOURLY_V*time.Minute)
	periodicTask(db.UpdateDayVisits, MINUTES_UPDATE_DAILY_V*time.Minute)
	periodicTask(db.UpdateMonthVisits, MINUTES_UPDATE_MONTHLY_V*time.Minute)
	periodicTask(db.UpdateHourDownloads, MINUTES_UPDATE_HOURLY_D*time.Minute)
	periodicTask(db.UpdateDayDownloads, MINUTES_UPDATE_DAILY_D*time.Minute)
	periodicTask(db.UpdateMonthDownloads, MINUTES_UPDATE_MONTHLY_D*time.Minute)
}

func periodicTask(task func() error, periodicity time.Duration) {
	go tasker(task, periodicity)
}

func tasker(task func() error, periodicity time.Duration) {
	for true {
		time.Sleep(periodicity)
		err := task()
		if err != nil {
			log.Error("Task error: ", err)
		}
	}
}
