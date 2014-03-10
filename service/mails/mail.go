package mails

import (
	"github.com/featen/ags/service/config"
	log "github.com/featen/utils/log"
	"net/http"
	"net/smtp"
)

var auth smtp.Auth

func SendMail(receiver string, subject string, message string) int {
	if auth == nil {
		auth = smtp.PlainAuth(
			"",
			config.GetValue("SenderEmail"),
			config.GetValue("SenderPassword"),
			config.GetValue("SmtpServer"))
	}

	log.Debug("auth is %v", auth)
	body := "To: " + receiver + "\r\nSubject: " + subject + "\r\n\r\n" + message
	err := smtp.SendMail(
		config.GetValue("SmtpServer")+":"+config.GetValue("SmtpPort"),
		auth,
		config.GetValue("SenderEmail"),
		[]string{receiver},
		[]byte(body))
	if err != nil {
		log.Info("mail %s sent to %s", subject, receiver)
		return http.StatusForbidden
	}
	return http.StatusOK
}

func SendRecoverMail(email, magic string) int {
	content := "You requested to recover your password via our website.\nIf you didn't do it, please ignore this mail.\nIf you confirm this recover, please click the below link. \nhttp://" + config.GetValue("Hostname") + "/service/recover/" + magic
	return SendMail(email, "Get back your access to "+config.GetValue("Hostname"), content)
}
