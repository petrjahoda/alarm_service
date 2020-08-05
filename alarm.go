package main

import (
	"database/sql"
	"github.com/petrjahoda/database"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Result string
}

func ProcessAlarm(alarm database.Alarm) {
	LogInfo(alarm.Name, "Processing alarm")
	timer := time.Now()
	alarmHasResult, alarmResult := ReadAlarmResult(alarm)
	alarmHasRecord := ReadAlarmRecord(alarm)
	if alarmHasResult && !alarmHasRecord {
		emailSent := SendAlarmEmail(alarm, alarmResult)
		if emailSent {
			CreateAlarmRecord(alarm)
		}
	} else if alarmHasRecord && !alarmHasResult {
		UpdateAlarmRecordToClosed(alarm)
	}
	LogInfo("MAIN", "Alarm processed in "+time.Since(timer).String())
}

func UpdateAlarmRecordToClosed(alarm database.Alarm) {
	LogInfo(alarm.Name, "Updating alarm record to closed")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(alarm.Name, "Problem opening database: "+err.Error())
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var alarmRecord database.AlarmRecord
	db.Where("alarm_id = ?", alarm.ID).Where("date_time_end is null").Find(&alarmRecord)
	alarmRecord.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	db.Save(&alarmRecord)
	LogInfo(alarm.Name, "Updated alarm record to closed in "+time.Since(timer).String())
}

func SendAlarmEmail(alarm database.Alarm, result string) bool {
	LogInfo(alarm.Name, "Sending alarm email")
	timer := time.Now()
	err, host, port, username, password, _ := ReadMailSettings(alarm)
	if err != nil {
		return false
	}
	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("Subject", alarm.MessageHeader)
	m.SetBody("text/html", alarm.MessageText+"\n\n"+result)
	UpdateRecipients(alarm, m)
	UpdateAttachments(alarm, m)
	d := gomail.NewDialer(host, port, username, password)
	if emailSentError := d.DialAndSend(m); emailSentError != nil {
		LogError(alarm.Name, "Email not sent: "+emailSentError.Error())
		return false
	} else {
		LogInfo(alarm.Name, "Email sent")
	}
	LogInfo(alarm.Name, "Sending alarm mail done, elapsed: "+time.Since(timer).String())
	return true
}

func UpdateAttachments(alarm database.Alarm, m *gomail.Message) {
	if len(alarm.Pdf) > 0 {
		CreatePdf(alarm)
		m.Attach(strconv.Itoa(int(alarm.ID)) + ".pdf")
	}
}

func UpdateRecipients(alarm database.Alarm, m *gomail.Message) {
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

func CreatePdf(alarm database.Alarm) {
	LogInfo(alarm.Name, "Creating pdf from "+alarm.Pdf)
	timer := time.Now()
	outputName := strconv.Itoa(int(alarm.ID)) + ".pdf"
	cmd := exec.Command("chromium-browser", "--headless", "--disable-gpu", "--no-sandbox", "--print-to-pdf="+outputName, alarm.Pdf)
	err := cmd.Run()
	if err != nil {
		LogError(alarm.Name, "Problem creating pdf: "+err.Error())
	}
	LogInfo(alarm.Name, "Pdf created in "+time.Since(timer).String())
}

func ReadMailSettings(alarm database.Alarm) (error, string, int, string, string, string) {
	LogInfo(alarm.Name, "Reading mail settings")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return nil, "", 0, "", "", ""
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var settingsHost database.Setting
	db.Where("name=?", "host").Find(&settingsHost)
	host := settingsHost.Value
	var settingsPort database.Setting
	db.Where("name=?", "port").Find(&settingsPort)
	port, err := strconv.Atoi(settingsPort.Value)
	if err != nil {
		LogError("MAIN", "Problem parsing port for email, using default port 587 "+err.Error())
		port = 587
	}
	var settingsUsername database.Setting
	db.Where("name=?", "username").Find(&settingsUsername)
	username := settingsUsername.Value
	var settingsPassword database.Setting
	db.Where("name=?", "password").Find(&settingsPassword)
	password := settingsPassword.Value
	var settingsEmail database.Setting
	db.Where("name=?", "email").Find(&settingsEmail)
	email := settingsEmail.Value
	LogInfo(alarm.Name, "Mail settings read in "+time.Since(timer).String())
	return err, host, port, username, password, email
}

func CreateAlarmRecord(alarm database.Alarm) {
	LogInfo(alarm.Name, "Creating alarm record")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(alarm.Name, "Problem opening database: "+err.Error())
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var alarmRecord database.AlarmRecord
	alarmRecord.DateTimeStart = time.Now()
	alarmRecord.AlarmID = int(alarm.ID)
	if alarm.WorkplaceID > 0 {
		alarmRecord.WorkplaceID = alarm.WorkplaceID
	}
	db.Save(&alarmRecord)
	LogInfo(alarm.Name, "Alarm record created in "+time.Since(timer).String())
}

func ReadAlarmRecord(alarm database.Alarm) bool {
	LogInfo(alarm.Name, "Reading alarm reacord")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(alarm.Name, "Problem opening database: "+err.Error())
		return false
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var alarmRecord database.AlarmRecord
	db.Where("alarm_id = ?", alarm.ID).Where("date_time_end is null").Find(&alarmRecord)
	alarmHasOpenRecord := alarmRecord.ID > 0
	if alarmHasOpenRecord {
		LogInfo(alarm.Name, "Alarm has open record")
		LogInfo(alarm.Name, "Alarm record read in "+time.Since(timer).String())
		return true
	}
	LogInfo(alarm.Name, "Alarm has not open record")
	LogInfo(alarm.Name, "Alarm record read in "+time.Since(timer).String())
	return false
}

func ReadAlarmResult(alarm database.Alarm) (bool, string) {
	LogInfo(alarm.Name, "Reading alarm results")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(alarm.Name, "Problem opening database: "+err.Error())
		return false, ""
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var result Result
	db.Raw(alarm.SqlCommand).Scan(&result)
	alarmHasResult := len(result.Result) > 0
	if !alarmHasResult {
		LogInfo(alarm.Name, "Alarm has no results")
		LogInfo(alarm.Name, "Alarm results read in "+time.Since(timer).String())
		return false, ""
	}
	LogInfo(alarm.Name, "Alarm has a result")
	LogInfo(alarm.Name, "Alarm results read in "+time.Since(timer).String())
	return true, result.Result
}
