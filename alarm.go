package main

import "github.com/petrjahoda/zapsi_database"

func ProcessAlarm(alarm zapsi_database.Alarm) {
	alarmHasResult, alarmResult := GetAlarmResult()
	alarmHasRecord := GetAlarmRecord()
	if alarmHasResult && !alarmHasRecord {
		CreateAlarmRecord(alarm)
		SendAlarmEmail(alarm, alarmResult)
	} else if alarmHasRecord && !alarmHasResult {
		CloseAlarmRecord(alarm)
	}

}

func CloseAlarmRecord(alarm zapsi_database.Alarm) {

}

func SendAlarmEmail(alarm zapsi_database.Alarm, result string) {

}

func CreateAlarmRecord(alarm zapsi_database.Alarm) {

}

func GetAlarmRecord() bool {
	return false
}

func GetAlarmResult() (bool, string) {
	return false, ""
}
