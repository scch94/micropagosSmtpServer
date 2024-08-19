package handler

import (
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scch94/ins_log"
	"github.com/scch94/micropagosSmtpServer/config"
)

func (handler *Handler) SendEmail(c *gin.Context) {

	ctx := c.Request.Context()
	ctx = ins_log.SetPackageNameInContext(ctx, "handler")

	ins_log.Info(ctx, "starting to send the email")

	configuration := config.Config

	smtpHost := configuration.SmtpData.SmtpHost
	smtpPort := configuration.SmtpData.SmtpPort
	senderEmail := configuration.MailInfo.MailSender
	subject := configuration.MailInfo.Subject
	hostname := configuration.SmtpData.HostName
	htmlBodyPath := configuration.MailInfo.UbicationMessage
	mailReceivers := configuration.MailInfo.MailReceivers
	//obtenemos los correos electronicos de los destinatarios esto vendra en el cuerpo o en una config por definir

	var recipients []string
	for i, receiver := range mailReceivers {
		ins_log.Infof(ctx, "%d: %s", i+1, receiver.Email)
		recipients = append(recipients, receiver.Email)
	}

	recipientEmails := strings.Join(recipients, ", ")

	// Establecer conexión con el servidor SMTP
	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		ins_log.Errorf(ctx, "Error when we try to connect with the SMTP server: %s ", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	defer client.Close()

	ins_log.Debug(ctx, "conectted to de smtp server")

	// Ejecutar el saludo (HELO)
	if err := client.Hello(hostname); err != nil {
		ins_log.Errorf(ctx, "Error when we try to specify the host name: %s ", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	ins_log.Debugf(ctx, "hostname %s", hostname)

	// Ejecutar el envío del remitente (MAIL FROM)
	if err := client.Mail(senderEmail); err != nil {
		ins_log.Errorf(ctx, "Failed to send MAIL FROM command: %s", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	ins_log.Debugf(ctx, "sender updated successfully")

	// Execute the RCPT TO command for all recipients
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			ins_log.Errorf(ctx, "Failed to deliver to recipient %s: %s", recipient, err)
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	}
	ins_log.Debugf(ctx, "recipients updated successfully")

	htmlBody, err := os.ReadFile(htmlBodyPath)
	if err != nil {
		ins_log.Errorf(ctx, "Error reading HTML body from %s: %s", htmlBodyPath, err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	// remplazamos los valores del html
	companyName := "Micropagos"
	dateYesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	body := strings.ReplaceAll(string(htmlBody), "[COMPANY_NAME]", companyName)
	body = strings.ReplaceAll(body, "[DATE_YESTERDAY]", dateYesterday)

	ins_log.Tracef(ctx, "HTML body has been created and placeholders have been replaced successfully")

	message := []byte(
		"To: " + recipientEmails + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: multipart/mixed; boundary=boundary\r\n" +
			"\r\n" +
			"--boundary\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
			"Content-Transfer-Encoding: 7bit\r\n" +
			"\r\n" +
			body + "\r\n" +
			"\r\n",
	)

	// Start writing the message (DATA)
	w, err := client.Data()
	if err != nil {
		ins_log.Errorf(ctx, "Error starting the message write: %s", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	defer w.Close()
	ins_log.Tracef(ctx, "Message write started successfully")

	// Write the message body
	if _, err := w.Write(message); err != nil {
		ins_log.Errorf(ctx, "Error writing the message body: %s", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	ins_log.Tracef(ctx, "Message body written successfully")

	// Finish writing the message
	if err := w.Close(); err != nil {
		ins_log.Errorf(ctx, "Error finishing the message write: %s", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	ins_log.Infof(ctx, "Message write finished successfully")

	// QUIT to end the SMTP session
	if err := client.Quit(); err != nil {
		ins_log.Errorf(ctx, "Error ending the SMTP session: %s", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	ins_log.Tracef(ctx, "SMTP session ended successfully")
	ins_log.Infof(ctx, "Email sent successfully!")

	c.JSON(http.StatusOK, nil)
}
