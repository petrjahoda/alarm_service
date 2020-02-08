package main

import (
	"github.com/jinzhu/gorm"
	"github.com/kardianos/service"
	"github.com/petrjahoda/zapsi_database"
	"strconv"
	"sync"
	"time"
)

const version = "2020.1.2.8"
const programName = "Alarm Service"
const programDesription = "Creates alarms for workplaces"
const deleteLogsAfter = 240 * time.Hour
const downloadInSeconds = 60

var (
	activeAlarms  []zapsi_database.Alarm
	runningAlarms []zapsi_database.Alarm
	alarmSync     sync.Mutex
)

type program struct{}

func (p *program) Start(s service.Service) error {
	LogInfo("MAIN", "Starting "+programName+" on "+s.Platform())
	go p.run()
	return nil
}

func (p *program) run() {
	LogDirectoryFileCheck("MAIN")
	LogInfo("MAIN", "Program version "+version+" started")
	CreateConfigIfNotExists()
	LoadSettingsFromConfigFile()
	LogDebug("MAIN", "Using ["+DatabaseType+"] on "+DatabaseIpAddress+":"+DatabasePort+" with database "+DatabaseName)
	for {
		start := time.Now()
		LogInfo("MAIN", "Program running")
		DeleteOldLogFiles()
		WriteProgramVersionIntoSettings()
		UpdateActiveAlarms("MAIN")
		LogInfo("MAIN", "Active alarms: "+strconv.Itoa(len(activeAlarms)))
		for _, activeAlarm := range activeAlarms {
			go RunAlarm(activeAlarm)

		}
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleeptime := downloadInSeconds*time.Second - time.Since(start)
			LogInfo("MAIN", "Sleeping for "+sleeptime.String())
			time.Sleep(sleeptime)
		}
	}
}
func (p *program) Stop(s service.Service) error {
	for len(runningAlarms) != 0 {
		LogInfo("MAIN", "Stopping, still running alarms: "+strconv.Itoa(len(runningAlarms)))
		time.Sleep(1 * time.Second)
	}
	LogInfo("MAIN", "Stopped on platform "+s.Platform())
	return nil
}

func main() {
	serviceConfig := &service.Config{
		Name:        programName,
		DisplayName: programName,
		Description: programDesription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		LogError("MAIN", err.Error())
	}
	err = s.Run()
	if err != nil {
		LogError("MAIN", "Problem starting "+serviceConfig.Name)
	}
}

func WriteProgramVersionIntoSettings() {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var settings zapsi_database.Setting
	db.Where("name=?", programName).Find(&settings)
	settings.Name = programName
	settings.Value = version
	db.Save(&settings)
	LogDebug("MAIN", "Updated version in database for "+programName)
}

func RunAlarm(alarm zapsi_database.Alarm) {
	LogInfo(alarm.Name, "Alarm started running")
	alarmSync.Lock()
	runningAlarms = append(runningAlarms, alarm)
	alarmSync.Unlock()
	start := time.Now()
	ProcessAlarm(alarm)
	LogInfo(alarm.Name, "Processing takes "+time.Since(start).String())
	RemoveAlarmFromRunningDevices(alarm)
	LogInfo(alarm.Name, "Alarm done, stopped running")

}

func RemoveAlarmFromRunningDevices(alarm zapsi_database.Alarm) {
	for idx, runningAlarm := range runningAlarms {
		if alarm.Name == runningAlarm.Name {
			alarmSync.Lock()
			runningAlarms = append(runningAlarms[0:idx], runningAlarms[idx+1:]...)
			alarmSync.Unlock()
		}
	}
}

func UpdateActiveAlarms(reference string) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(reference, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeAlarms = nil
		return
	}
	defer db.Close()
	db.Find(&activeAlarms)
}
