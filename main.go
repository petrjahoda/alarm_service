package main

import (
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"strconv"
	"sync"
	"time"
)

const version = "2021.1.2.21"
const serviceName = "Alarm Service"
const serviceDescription = "Creates alarms for workplaces"
const downloadInSeconds = 60
const config = "user=postgres password=pj79.. dbname=system host=database port=5432 sslmode=disable"

var (
	activeAlarms  []database.Alarm
	runningAlarms []database.Alarm
	alarmSync     sync.Mutex
)

type program struct{}

func main() {
	logInfo("MAIN", serviceName+" ["+version+"] starting...")
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		logError("MAIN", "Cannot start: "+err.Error())
	}
	err = s.Run()
	if err != nil {
		logError("MAIN", "Cannot start: "+err.Error())
	}
}

func (p *program) Start(service.Service) error {
	logInfo("MAIN", serviceName+" ["+version+"] started")
	go p.run()
	return nil
}

func (p *program) Stop(service.Service) error {
	for len(runningAlarms) != 0 {
		logInfo("MAIN", serviceName+" ["+version+"] stopping...")
		time.Sleep(1 * time.Second)
	}
	logInfo("MAIN", serviceName+" ["+version+"] stopped")
	return nil
}

func (p *program) run() {
	updateProgramVersion()
	for {
		logInfo("MAIN", serviceName+" ["+version+"] running")
		start := time.Now()
		readActiveAlarms("MAIN")
		logInfo("MAIN", "Active alarms: "+strconv.Itoa(len(activeAlarms)))
		for _, activeAlarm := range activeAlarms {
			go runAlarm(activeAlarm)

		}
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleepTime := downloadInSeconds*time.Second - time.Since(start)
			logInfo("MAIN", "Sleeping for "+sleepTime.String())
			time.Sleep(sleepTime)
		}
	}
}
