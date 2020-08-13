package main

import (
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"strconv"
	"sync"
	"time"
)

const version = "2020.3.2.13"
const serviceName = "Alarm Service"
const serviceDescription = "Creates alarms for workplaces"
const downloadInSeconds = 60
const config = "user=postgres password=Zps05..... dbname=version3 host=database port=5432 sslmode=disable"

var (
	activeAlarms  []database.Alarm
	runningAlarms []database.Alarm
	alarmSync     sync.Mutex
)

type program struct{}

func main() {
	LogInfo("MAIN", serviceName+" ["+version+"] starting...")
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		LogError("MAIN", "Cannot start: "+err.Error())
	}
	err = s.Run()
	if err != nil {
		LogError("MAIN", "Cannot start: "+err.Error())
	}
}

func (p *program) Start(service.Service) error {
	LogInfo("MAIN", serviceName+" ["+version+"] started")
	go p.run()
	return nil
}

func (p *program) Stop(service.Service) error {
	for len(runningAlarms) != 0 {
		LogInfo("MAIN", serviceName+" ["+version+"] stopping...")
		time.Sleep(1 * time.Second)
	}
	LogInfo("MAIN", serviceName+" ["+version+"] stopped")
	return nil
}

func (p *program) run() {
	UpdateProgramVersion()
	for {
		LogInfo("MAIN", serviceName+" ["+version+"] running")
		start := time.Now()
		ReadActiveAlarms("MAIN")
		LogInfo("MAIN", "Active alarms: "+strconv.Itoa(len(activeAlarms)))
		for _, activeAlarm := range activeAlarms {
			go RunAlarm(activeAlarm)

		}
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleepTime := downloadInSeconds*time.Second - time.Since(start)
			LogInfo("MAIN", "Sleeping for "+sleepTime.String())
			time.Sleep(sleepTime)
		}
	}
}
