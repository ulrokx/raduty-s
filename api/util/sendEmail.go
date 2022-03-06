package util

import (
	"fmt"
	"net/smtp"
)

var emailAuth smtp.Auth

func SendEmailSMTP(to []string, subject string) (bool, error) {
	emailHost := "smtp.gmail.com"
	emailFrom := "stevensrassistants@gmail.com"
	emailPassword := "D32gKfZq7e2q"
	emailPort := 587

	emailAuth = smtp.PlainAuth("", emailFrom, emailPassword, emailHost)

	emailBody := "Thank you for registering your availability!"

	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	subjectline := "Subject: " + subject + "!\n"
	msg := []byte(subjectline + mime + "\n" + emailBody)
	addr := fmt.Sprintf("%s:%d", emailHost, emailPort)

	if err := smtp.SendMail(addr, emailAuth, emailFrom, to, msg); err != nil {
		return false, err
	}
	return true, nil
}
