package main

import (
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

const version = "2020.3.1.30"
const programName = "Alarm Service"
const programDescription = "Creates alarms for workplaces"
const downloadInSeconds = 60
const config = "user=postgres password=Zps05..... dbname=version3 host=database port=5432 sslmode=disable"

var (
	activeAlarms  []database.Alarm
	runningAlarms []database.Alarm
	alarmSync     sync.Mutex
)

type program struct{}

func (p *program) Start(s service.Service) error {
	LogInfo("MAIN", "Starting "+programName+" on "+s.Platform())
	go p.run()
	return nil
}

func (p *program) run() {
	LogInfo("MAIN", "Program version "+version+" started")
	WriteProgramVersionIntoSettings()
	for {
		start := time.Now()
		LogInfo("MAIN", "Program running")
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
		Description: programDescription,
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
	LogInfo("MAIN", "Updating program version in database")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var settings database.Setting
	db.Where("name=?", programName).Find(&settings)
	settings.Name = programName
	settings.Value = version
	db.Save(&settings)
	LogInfo("MAIN", "Program version updated, elapsed: "+time.Since(timer).String())
}

func RunAlarm(alarm database.Alarm) {
	LogInfo(alarm.Name, "Alarm loop started")
	timer := time.Now()
	alarmSync.Lock()
	runningAlarms = append(runningAlarms, alarm)
	alarmSync.Unlock()
	ProcessAlarm(alarm)
	RemoveAlarmFromRunningDevices(alarm)
	LogInfo("MAIN", "Alarm loop ended, elapsed: "+time.Since(timer).String())

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

func UpdateActiveAlarms(reference string) {
	LogInfo("MAIN", "Updating active alarms")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(reference, "Problem opening  database: "+err.Error())
		activeAlarms = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	db.Find(&activeAlarms)
	LogInfo("MAIN", "Active alarms updated, elapsed: "+time.Since(timer).String())

}
