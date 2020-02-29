package main

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/petrjahoda/zapsi_database"
	"gopkg.in/gomail.v2"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Result string
}

func ProcessAlarm(alarm zapsi_database.Alarm) {
	LogInfo(alarm.Name, "Processing alarm")
	timer := time.Now()
	alarmHasResult, alarmResult := CheckAlarmResult(alarm)
	alarmHasRecord := CheckAlarmRecord(alarm)
	if alarmHasResult && !alarmHasRecord {
		CreateAlarmRecord(alarm)
		SendAlarmEmail(alarm, alarmResult)
	} else if alarmHasRecord && !alarmHasResult {
		CloseAlarmRecord(alarm)
	} else if alarmHasResult && alarmHasRecord {
		UpdateAlarm(alarm)
	}
	LogInfo("MAIN", "Alarm processed, elapsed: "+time.Since(timer).String())

}

func UpdateAlarm(alarm zapsi_database.Alarm) {
	LogInfo(alarm.Name, "Updating alarm record")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(alarm.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
	}
	defer db.Close()
	var alarmRecord zapsi_database.AlarmRecord
	db.Where("alarm_id = ?", alarm.ID).Where("date_time_end is null").Find(&alarmRecord)
	alarmRecord.Duration = time.Now().Sub(alarmRecord.DateTimeStart)
	db.Save(&alarmRecord)
	LogInfo(alarm.Name, "Updating alarm record done, elapsed: "+time.Since(timer).String())

}

func CloseAlarmRecord(alarm zapsi_database.Alarm) {
	LogInfo(alarm.Name, "Closing alarm record")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(alarm.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
	}
	defer db.Close()
	var alarmRecord zapsi_database.AlarmRecord
	db.Where("alarm_id = ?", alarm.ID).Where("date_time_end is null").Find(&alarmRecord)
	alarmRecord.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	alarmRecord.Duration = time.Now().Sub(alarmRecord.DateTimeStart)
	db.Save(&alarmRecord)
	LogInfo(alarm.Name, "Closing alarm record done, elapsed: "+time.Since(timer).String())

}

func SendAlarmEmail(alarm zapsi_database.Alarm, result string) {
	LogInfo(alarm.Name, "Sending alarm email")
	timer := time.Now()
	err, host, port, username, password, _ := UpdateMailSettings()
	if err != nil {
		return
	}
	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("Subject", alarm.MessageHeader)
	m.SetBody("text/html", alarm.MessageText+"\n\n"+result)
	UpdateRecipients(alarm, m)
	UpdateAttachements(alarm, m)

	d := gomail.NewDialer(host, port, username, password) // PETRzpsMAIL79..
	if emailSentError := d.DialAndSend(m); emailSentError != nil {
		LogError(alarm.Name, "Email not sent: "+emailSentError.Error())
	} else {
		LogInfo(alarm.Name, "Email sent")
	}
	LogInfo(alarm.Name, "Sending alarm mail done, elapsed: "+time.Since(timer).String())

}

func UpdateAttachements(alarm zapsi_database.Alarm, m *gomail.Message) {
	if len(alarm.Pdf) > 0 {
		ConvertToPdf(alarm)
		m.Attach(strconv.Itoa(int(alarm.ID)) + ".pdf")
	}
}

func UpdateRecipients(alarm zapsi_database.Alarm, m *gomail.Message) {
	if strings.Contains(alarm.Recipients, ",") {
		emails := strings.Split(alarm.Recipients, ",")
		m.SetHeader("To", emails...)
	} else if strings.Contains(alarm.Recipients, ";") {
		emails := strings.Split(alarm.Recipients, ";")
		m.SetHeader("To", emails...)
	} else {
		m.SetHeader("To", alarm.Recipients)
	}
}

func ConvertToPdf(alarm zapsi_database.Alarm) {
	LogInfo(alarm.Name, "Creating pdf from "+alarm.Pdf)
	outputName := strconv.Itoa(int(alarm.ID)) + ".pdf"
	cmd := exec.Command("/usr/bin/chromium-browser", "--headless", "--disable-gpu", "--no-sandbox", "--print-to-pdf="+outputName, alarm.Pdf)
	err := cmd.Run()
	if err != nil {
		LogError(alarm.Name, "Problem creating pdf: "+err.Error())
	}
	LogInfo(alarm.Name, "Pdf created")
}

func UpdateMailSettings() (error, string, int, string, string, string) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return nil, "", 0, "", "", ""
	}
	var settingsHost zapsi_database.Setting
	db.Where("name=?", "host").Find(&settingsHost)
	host := settingsHost.Value
	var settingsPort zapsi_database.Setting
	db.Where("name=?", "port").Find(&settingsPort)
	port, err := strconv.Atoi(settingsPort.Value)
	if err != nil {
		LogError("MAIN", "Problem parsing port for email, using default port 587 "+err.Error())
		port = 587
	}
	var settingsUsername zapsi_database.Setting
	db.Where("name=?", "username").Find(&settingsUsername)
	username := settingsUsername.Value
	var settingsPassword zapsi_database.Setting
	db.Where("name=?", "password").Find(&settingsPassword)
	password := settingsPassword.Value
	var settingsEmail zapsi_database.Setting
	db.Where("name=?", "email").Find(&settingsEmail)
	email := settingsEmail.Value
	defer db.Close()
	return err, host, port, username, password, email
}

func CreateAlarmRecord(alarm zapsi_database.Alarm) {
	LogInfo(alarm.Name, "Checking alarm record")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(alarm.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
	}
	defer db.Close()
	var alarmRecord zapsi_database.AlarmRecord
	alarmRecord.DateTimeStart = time.Now()
	alarmRecord.Duration = 0
	alarmRecord.AlarmId = alarm.ID
	if alarm.WorkplaceId > 0 {
		alarmRecord.WorkplaceId = sql.NullInt32{Int32: int32(alarm.WorkplaceId), Valid: true}
	}
	db.Save(&alarmRecord)
	LogInfo(alarm.Name, "Alarm record creating done, elapsed: "+time.Since(timer).String())

}

func CheckAlarmRecord(alarm zapsi_database.Alarm) bool {
	LogInfo(alarm.Name, "Checking open alarm record")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(alarm.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		return false
	}
	defer db.Close()
	var alarmRecord zapsi_database.AlarmRecord
	db.Where("alarm_id = ?", alarm.ID).Where("date_time_end is null").Find(&alarmRecord)
	alarmHasOpenRecord := alarmRecord.ID > 0
	LogInfo(alarm.Name, "Checking alarm record done, elapsed: "+time.Since(timer).String())
	if alarmHasOpenRecord {
		return true
	}
	return false
}

func CheckAlarmResult(alarm zapsi_database.Alarm) (bool, string) {
	LogInfo(alarm.Name, "Checking alarm results")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(alarm.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		return false, ""
	}
	defer db.Close()
	var result Result
	db.Raw(alarm.SqlCommand).Scan(&result)
	alarmHasResult := len(result.Result) > 0
	if !alarmHasResult {
		LogInfo(alarm.Name, "Alarm has no results")
		LogInfo(alarm.Name, "Checking alarm results done, elapsed: "+time.Since(timer).String())
		return false, ""

	}
	LogInfo(alarm.Name, "Alarm has a result")
	LogInfo(alarm.Name, "Checking alarm results done, elapsed: "+time.Since(timer).String())
	return true, result.Result
}
