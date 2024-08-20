package handler

import (
	"context"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scch94/ins_log"
	"github.com/scch94/micropagosSmtpServer/config"
	"github.com/scch94/micropagosSmtpServer/models/request"
	"github.com/scch94/micropagosSmtpServer/models/response"
)

const (
	STATUSOK      = "0"
	FORWARDREF    = "3"
	DESCRIPTIONOK = "No errors"
	STATUSERROR   = "5"
)

func (handler *Handler) SendEmail(c *gin.Context) {

	ctx := c.Request.Context()
	ctx = ins_log.SetPackageNameInContext(ctx, "handler")

	//guardamos la info del emailRequest
	var emailRequest request.SendEmailRequest
	if err := c.BindJSON(&emailRequest); err != nil {
		ins_log.Errorf(ctx, "error when we try to get the json petition")
		response := response.NewSendEmailResponse(STATUSERROR, FORWARDREF, err.Error())
		c.JSON(http.StatusOK, response)
		return
	}

	//setiamos el utfi para poder hacer trasa
	ctx = ins_log.SetUTFIInContext(ctx, emailRequest.Utfi)
	ins_log.Info(ctx, "new petition to send email received")

	//generamos el html q sera enviado
	formaterBody, err := generateHTML(ctx, emailRequest)
	if err != nil {
		ins_log.Errorf(ctx, "Error in the function generateHTML() err:%v", err)
		response := response.NewSendEmailResponse(STATUSERROR, FORWARDREF, err.Error())
		c.JSON(http.StatusOK, response)
		return
	}

	//creamos el servidor smtp q usaremos apra enviar el correo
	client, err := getSmtpClient(ctx, emailRequest)
	if err != nil {
		ins_log.Errorf(ctx, "Error in the function getSmtpClient()  err:%v", err)
		response := response.NewSendEmailResponse(STATUSERROR, FORWARDREF, err.Error())
		c.JSON(http.StatusOK, response)
		return
	}

	//enviaremos el correo y cerraremos el smtpclient
	err = writeAndSendEmail(ctx, formaterBody, client)
	if err != nil {
		ins_log.Errorf(ctx, "Error in the function writeAndSendEmail() err:%v", err)
		response := response.NewSendEmailResponse(STATUSERROR, FORWARDREF, err.Error())
		c.JSON(http.StatusOK, response)
		return
	}

	ins_log.Infof(ctx, "Email sent successfully!")

	response := response.NewSendEmailResponse(STATUSOK, FORWARDREF, DESCRIPTIONOK)

	c.JSON(http.StatusOK, response)
}

func getSmtpClient(ctx context.Context, requestEmail request.SendEmailRequest) (*smtp.Client, error) {

	//absorvemos valores de configuracion para crear el client de smtp
	configuration := config.Config
	smtpHost := configuration.SmtpData.SmtpHost
	smtpPort := configuration.SmtpData.SmtpPort
	hostname := configuration.SmtpData.HostName
	senderEmail := configuration.MailInfo.MailSender

	ins_log.Infof(ctx, "starting to create the smtp client smtpHost: %s, smtpPort: %s", smtpHost, smtpPort)

	//creamos el client
	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		ins_log.Errorf(ctx, "Error when we try to connect with the SMTP server: %s ", err)
		return nil, err
	}

	ins_log.Debug(ctx, "conectted to de smtp server")

	// Ejecutar el saludo (HELO)
	if err := client.Hello(hostname); err != nil {
		ins_log.Errorf(ctx, "Error when we try to specify the host name: %s ", err)
		return nil, err
	}

	ins_log.Tracef(ctx, "hostname %s", hostname)

	// Ejecutar el envío del remitente (MAIL FROM)
	if err := client.Mail(senderEmail); err != nil {
		ins_log.Errorf(ctx, "Failed to send MAIL FROM command: %s", err)
		return nil, err
	}

	ins_log.Tracef(ctx, "the source email address is %s", senderEmail)

	//configuramos el correo a las personas a las que se lo vamos a enviar
	recipients := configuration.GetEmailsToSend(ctx, requestEmail.Client)
	recipientEmails := strings.Join(recipients, ", ")
	// Execute the RCPT TO command for all recipients
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			ins_log.Errorf(ctx, "Failed to deliver to recipient %s: %s", recipient, err)
			return nil, err
		}
	}
	ins_log.Tracef(ctx, "recipients updated successfully recipients: %s", recipientEmails)
	return client, nil
}

func generateHTML(ctx context.Context, requestEmail request.SendEmailRequest) ([]byte, error) {

	ins_log.Infof(ctx, "starting to generate HTML mail")

	//recuperamos los valores del config que usaremos
	configuration := config.Config
	htmlBodyPath := configuration.MailInfo.UbicationMessage

	// recuperamos los valores del request que usaremos
	origen := requestEmail.OriginNumber
	destino := requestEmail.GetDestination()
	telco := requestEmail.TLVValue
	subject := configuration.MailInfo.Subject + destino

	//traemos los correos a los que enviaremos el correo y generamos el strin concatenado que usaremos en formatedbody como to
	recipients := configuration.GetEmailsToSend(ctx, requestEmail.Client)
	recipientEmails := strings.Join(recipients, ", ")

	//tremos el contenido del mensaje
	message, err := requestEmail.GetMessage(ctx)
	if err != nil {
		ins_log.Errorf(ctx, "error in the function getMessage(): %v", err)
		return nil, err
	}

	ins_log.Infof(ctx, "this is the message to send: %v", message)
	ins_log.Tracef(ctx, "this is the subject :%s", subject)

	//leemos el html seung la ubicacion dicha en el archivo de configuracion
	htmlBody, err := os.ReadFile(htmlBodyPath)
	if err != nil {
		ins_log.Errorf(ctx, "Error reading HTML body from %s: %s", htmlBodyPath, err)
		return nil, err
	}

	//remplazamos los valores del html
	body := strings.ReplaceAll(string(htmlBody), "[TELCO_NAME]", telco)
	body = strings.ReplaceAll(body, "[ORIGIN]", origen)
	body = strings.ReplaceAll(body, "[DESTINITY]", destino)
	body = strings.ReplaceAll(body, "[MESSAGE]", message)

	ins_log.Tracef(ctx, "HTML body has been created and placeholders have been replaced successfully")

	formaterBody := []byte(
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

	return formaterBody, nil
}

func writeAndSendEmail(ctx context.Context, formaterBody []byte, client *smtp.Client) error {
	defer client.Close()
	ins_log.Infof(ctx, "starting to send email ")

	// Start writing the message (DATA)
	w, err := client.Data()
	if err != nil {
		ins_log.Errorf(ctx, "Error starting the message write: %s", err)
		return err
	}
	defer w.Close()
	ins_log.Tracef(ctx, "Message write started successfully")

	// Write the message body
	if _, err := w.Write(formaterBody); err != nil {
		ins_log.Errorf(ctx, "Error writing the message body: %s", err)
		return err
	}

	ins_log.Tracef(ctx, "Message body written successfully")

	// Finish writing the message
	if err := w.Close(); err != nil {
		ins_log.Errorf(ctx, "Error finishing the message write: %s", err)
		return err
	}
	ins_log.Infof(ctx, "Message write finished successfully")

	// QUIT to end the SMTP session
	if err := client.Quit(); err != nil {
		ins_log.Errorf(ctx, "Error ending the SMTP session: %s", err)
		return err
	}
	ins_log.Tracef(ctx, "SMTP session ended successfully")
	return nil
}
