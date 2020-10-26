package main

import (
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func updateProgramVersion() {
	logInfo("MAIN", "Writing program version into settings")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	var existingSettings database.Setting
	db.Where("name=?", serviceName).Find(&existingSettings)
	existingSettings.Name = serviceName
	existingSettings.Value = version
	db.Save(&existingSettings)
	logInfo("MAIN", "Program version written into settings in "+time.Since(timer).String())
}

func runAlarm(alarm database.Alarm) {
	logInfo(alarm.Name, "Alarm main loop started")
	timer := time.Now()
	alarmSync.Lock()
	runningAlarms = append(runningAlarms, alarm)
	alarmSync.Unlock()
	processAlarm(alarm)
	removeAlarmFromRunningDevices(alarm)
	logInfo("MAIN", "Alarm main loop ended in "+time.Since(timer).String())

}

func removeAlarmFromRunningDevices(alarm database.Alarm) {
	for idx, runningAlarm := range runningAlarms {
		if alarm.Name == runningAlarm.Name {
			alarmSync.Lock()
			runningAlarms = append(runningAlarms[0:idx], runningAlarms[idx+1:]...)
			alarmSync.Unlock()
		}
	}
}

func readActiveAlarms(reference string) {
	logInfo("MAIN", "Reading active alarms")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(reference, "Problem opening database: "+err.Error())
		activeAlarms = nil
		return
	}
	db.Find(&activeAlarms)
	logInfo("MAIN", "Active alarms read in "+time.Since(timer).String())

}
