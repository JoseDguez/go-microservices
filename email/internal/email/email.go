package email

import (
	"fmt"
	"net/smtp"
)

func Send(target string, orderID string) error {
	senderEmail := "email@gmail.com"
	senderPassword := "password"

	recipientEmail := target

	message := []byte(fmt.Sprintf("Subject: Payment Processed!\nProcess ID: %s\n", orderID))

	smtpServer := "smtp.gmail.com"
	smtpPort := 587

	creds := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)

	smtpAddress := fmt.Sprintf("%s:%d", smtpServer, smtpPort)
	err := smtp.SendMail(smtpAddress, creds, senderEmail, []string{recipientEmail}, message)
	if err != nil {
		return err
	}

	return nil
}
