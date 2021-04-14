package main

import(
	"log"
	gomail "github.com/xhit/go-simple-mail"
	"strconv"
	"os"
	"fmt"
)
func connectToMailServer()(*gomail.SMTPClient, error){
	client := gomail.NewSMTPClient()
	if os.Getenv("SMTP_HOST") == "" {
		return nil, fmt.Errorf("[ENV] SMTP_HOST missing!")
	}else if os.Getenv("SMTP_PORT") == "" {
		return nil, fmt.Errorf("[ENV] SMTP_PORT missing!")
	}else if os.Getenv("SMTP_USER") == "" {
		return nil, fmt.Errorf("[ENV] SMTP_USER missing!")
	}else if os.Getenv("SMTP_PASSWORD") == "" {
		return nil, fmt.Errorf("[ENV] SMTP_PASSWORD missing!")
	}
	client.Host = os.Getenv("SMTP_HOST")
	client.Port, _ = strconv.Atoi(os.Getenv("SMTP_PORT"))
	client.Username = os.Getenv("SMTP_USER")
	client.Password = os.Getenv("SMTP_PASSWORD")
	client.Encryption = gomail.EncryptionSSL

	return client.Connect()
}
func sendMailSimple(server system, sendRequest mail, mc *MailCollector) error {
	if sendRequest.From == "" {
		return fmt.Errorf("[MAIL] From can not be empty")
	}else if sendRequest.To == "" {
		return fmt.Errorf("[MAIL] To can not be empty")
	}else if sendRequest.Subject == "" {
		return fmt.Errorf("[MAIL] Subject can not be empty")
	}else if os.Getenv("MAIL_OFF") == "true"{
		log.Println("[MAIL] Mailversand ist deaktiviert (MAIL_OFF == true)")
		return nil
	}

	log.Printf("[MAIL] \"%s\" for \"%s\" from \"%s\"\n", sendRequest.Subject, sendRequest.To, sendRequest.From)

	smtpClient, err := connectToMailServer()
	if err != nil {
		log.Fatalf("[MAIL] %s\n", err) 
	}

	m := gomail.NewMSG()
	m.SetFrom(sendRequest.From)
	m.SetReplyTo(sendRequest.From)
	m.AddTo(sendRequest.To)
	if os.Getenv("MAIL_BCC") == "true" {
		m.AddBcc(sendRequest.From)
		if m.Error != nil{
			log.Printf("[MAIL] found an error, but deleting it, as it is possible that BCC = TO\n")
			log.Printf("[MAIL] BCC = %s | TO = %s\n", sendRequest.From, sendRequest.To)
			log.Printf("[MAIL] Error: %s\n", m.Error)
			m.Error = nil
		}
	}
	if sendRequest.MimeType == "text/html" {
		log.Printf("[MAIL] Setting ContentType to text/html\n")
		m.SetBody(gomail.TextHTML, sendRequest.Body)
	}else{
		//Default is plain
		log.Printf("[MAIL] Setting ContentType to text/plain\n")
		m.SetBody(gomail.TextPlain, sendRequest.Body)
	}
	m.SetSubject(sendRequest.Subject)

	for _, a := range sendRequest.Attachments.Results{
		if a.correctLoaded != true {
			log.Printf("[MAIL] Ignoring Attachment %s with Partnumber %d", a.Name, a.Partno)
			continue
		}
		m.AddAttachmentData(a.data, a.Name, a.MimeType)
		log.Printf("[MAIL] Added Attachment %s with Partnumber %d to Mail", a.Name, a.Partno)

	}
	err = m.SendEnvelopeFrom(os.Getenv("SMTP_USER"), smtpClient)

	if err != nil{
		return fmt.Errorf("[MAIL] %s", err)
	}
	err = m.GetError()
	if err != nil {
		return fmt.Errorf("[MAIL] %s", err)
	}
	mc.incMails(server.SystemName, server.Mandant)
	return nil
}
func sendMailError(server system, logs []string, mc *MailCollector)error{
	log.Println("[MAIL] Generating Error Mail from Logs")
	mail := mail{}
	mail.To = os.Getenv("ERROR_MAIL")
	mail.From = os.Getenv("ERROR_MAIL")
	mail.Subject = fmt.Sprintf("%s %s[%s]", os.Getenv("ERROR_SUBJECT"), server.SystemName, server.Mandant)

	for _, logEntry := range logs {
		mail.Body += logEntry + "\n"
	}

	log.Printf("[MAIL] Sending Errormail to %s with subject %s\n", mail.To, mail.Subject)
	mc.incErrMails(server.SystemName, server.Mandant)

	return sendMailSimple(server, mail, mc)
}