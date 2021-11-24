package main

import(
	"log"
	gomail "github.com/xhit/go-simple-mail/v2"
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
	if client.Port == 465 {
		log.Println("[MAIL] Using Encryption SSL")
		client.Encryption = gomail.EncryptionSSLTLS
	}else if client.Port == 587 {
		log.Println("[MAIL] Using Encryption STARTSSL")
		client.Encryption = gomail.EncryptionSTARTTLS
	}else{
		return nil, fmt.Errorf("[ENV] PORT is set to an unknown int, must be 465 or 587")
	}
	if client.Host == "smtp.office365.com" {
		log.Println("[MAIL] Detected office365, using Authentication \"AuthLogin\"")
		client.Authentication = gomail.AuthLogin
	}
	
	return client.Connect()
}
func sendMailSimple(server system, sendRequest mail, mc *MailCollector) error {
	if sendRequest.From == "" {
		return fmt.Errorf("[MAIL] From can not be empty")
	}else if sendRequest.Subject == "" {
		return fmt.Errorf("[MAIL] Subject can not be empty")
	}else if os.Getenv("MAIL_OFF") == "true"{
		log.Println("[MAIL] Mailversand ist deaktiviert (MAIL_OFF == true)")
		return nil
	}

	if os.Getenv("SEND_AS_FORBIDDEN") != "" && os.Getenv("SMTP_FROM") != sendRequest.From {
		log.Printf("[MAIL] SEND_AS is forbidden by customizing, replacing Sender %s with %s", sendRequest.From, os.Getenv("SMTP_FROM"))
		sendRequest.From = os.Getenv("SMTP_FROM")
		
	}

	log.Printf("[MAIL] \"%s\" for \"%s\" from \"%s\"\n", sendRequest.Subject, sendRequest.To, sendRequest.From)

	smtpClient, err := connectToMailServer()
	if err != nil {
		log.Fatalf("[MAIL] %s\n", err) 
	}

	m := gomail.NewMSG()
	m.SetFrom(sendRequest.From)
	m.SetReplyTo(sendRequest.From)
	addedRecipient := false
	for _, receiver := range sendRequest.Receivers.Results{
		log.Printf("[MAIL] Adding receiver %s in list for Mode %s", receiver.Mail, receiver.Modus)
		if receiver.Modus == "BCC" {
			m.AddBcc(receiver.Mail)
		}else if receiver.Modus == "CC" {
			m.AddCc(receiver.Mail)
		}else {
			m.AddTo(receiver.Mail)
		}
		addedRecipient = true
	}
	if sendRequest.To != "" {
		m.AddTo(sendRequest.To)
		addedRecipient = true
	}
	if ! addedRecipient {
		log.Printf("[MAIL] No recipients?")
		return fmt.Errorf("[MAIL] No recipient found")
	}
	if server.ActivateBCC {
		log.Println("[MAIL] BCC is activated")
		m.AddBcc(sendRequest.From)
		if m.Error != nil{
			log.Printf("[MAIL] BCC = %s | TO = %s\n", sendRequest.From, sendRequest.To)
			log.Printf("[MAIL] Error: %s\n", m.Error)
		}
	}
	if m.Error != nil{
		return fmt.Errorf("[MAIL] Error: %s", m.Error)
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
	mail.From = os.Getenv("SMTP_FROM")
	mail.Subject = fmt.Sprintf("%s %s[%s]", os.Getenv("ERROR_SUBJECT"), server.SystemName, server.Mandant)

	for _, logEntry := range logs {
		mail.Body += logEntry + "\n"
	}

	log.Printf("[MAIL] Sending Errormail to %s with subject %s\n", mail.To, mail.Subject)
	mc.incErrMails(server.SystemName, server.Mandant)

	return sendMailSimple(server, mail, mc)
}
