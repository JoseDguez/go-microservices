package email

import (
	"fmt"
	"net/smtp"
	"os"
)

func Send(target string, orderID string) error {
	senderEmail := os.Getenv("SENDER_EMAIL")
	senderPassword := os.Getenv("SENDER_PASSWORD")

	recipientEmail := target

	message := []byte(fmt.Sprintf("Subject: Payment Processed!\nProcess ID: %s\n", orderID))

	smtpServer := os.Getenv("EMAIL_HOST")
	smtpPort := os.Getenv("EMAIL_PORT")

	credentials := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)

	smtpAddress := fmt.Sprintf("%s:%s", smtpServer, smtpPort)
	err := smtp.SendMail(smtpAddress, credentials, senderEmail, []string{recipientEmail}, message)
	if err != nil {
		return err
	}

	return nil
}
