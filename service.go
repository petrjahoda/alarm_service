package main

import (
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func UpdateProgramVersion() {
	LogInfo("MAIN", "Writing program version into settings")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var existingSettings database.Setting
	db.Where("name=?", serviceName).Find(&existingSettings)
	existingSettings.Name = serviceName
	existingSettings.Value = version
	db.Save(&existingSettings)
	LogInfo("MAIN", "Program version written into settings in "+time.Since(timer).String())
}

func RunAlarm(alarm database.Alarm) {
	LogInfo(alarm.Name, "Alarm main loop started")
	timer := time.Now()
	alarmSync.Lock()
	runningAlarms = append(runningAlarms, alarm)
	alarmSync.Unlock()
	ProcessAlarm(alarm)
	RemoveAlarmFromRunningDevices(alarm)
	LogInfo("MAIN", "Alarm main loop ended in "+time.Since(timer).String())

}

func RemoveAlarmFromRunningDevices(alarm database.Alarm) {
	for idx, runningAlarm := range runningAlarms {
		if alarm.Name == runningAlarm.Name {
			alarmSync.Lock()
			runningAlarms = append(runningAlarms[0:idx], runningAlarms[idx+1:]...)
			alarmSync.Unlock()
		}
	}
}

func ReadActiveAlarms(reference string) {
	LogInfo("MAIN", "Reading active alarms")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(reference, "Problem opening database: "+err.Error())
		activeAlarms = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	db.Find(&activeAlarms)
	LogInfo("MAIN", "Active alarms read in "+time.Since(timer).String())

}
