package main

import (
	"log"
	"net/http"
	"fmt"
	"encoding/json"
	"time"
	"os"
	"io/ioutil"
)
type mail struct {
	Objtp string `json:"Objtp"`
	Objyr string `json:"Objyr"`
	Objno string `json:"Objno"`
	Fortp string `json:"Fortp"`
	Foryr string `json:"Foryr"`
	Forno string `json:"Forno"`
	Rectp string `json:"Rectp"`
	Recyr string `json:"Recyr"`
	Recno string `json:"Recno"`
	MimeType string `json:"MimeType"`

	To string `json:"Mailto"`
	From string `json:"Mailfrom"`
	Subject string `json:"Subject"`
	Body string `json:"Body"`
	Attachments attachmentResults `json:"Attachments"`
}
type attachmentResults struct {
	Results []attachment `json:"results"`
}
type attachment struct {
	Objtp string `json:"Objtp"`
	Objyr string `json:"Objyr"`
	Objno string `json:"Objno"`
	Fortp string `json:"Fortp"`
	Foryr string `json:"Foryr"`
	Forno string `json:"Forno"`
	Rectp string `json:"Rectp"`
	Recyr string `json:"Recyr"`
	Recno string `json:"Recno"`
	Partno int `json:"Partno"`

	Name string `json:"Name"`
	Docsize int `json:"Docsize"`
	Doctype string `json:"Doctype"`
	MimeType string `json:"Mimetype"`
	data []byte
	correctLoaded bool
}
type result struct {
	Results []mail `json:"results"`
}
type odata struct {
	Data result `json:"d"`
}
func doSystem(server system, mc *MailCollector){
	log.Printf("Checking System \"%s\" with mandant \"%s\"\n", server.SystemName, server.Mandant)
	timeStart :=  time.Now()
	err := getMailFromServer(server, mc)
	mc.setLastLoopDuration(server.SystemName, server.Mandant, time.Since(timeStart))
	if err != nil {
		log.Printf("Error checking %s[%s]: %s\n", server.SystemName, server.Mandant, err)
	}
	log.Printf("Checking System \"%s\" with mandant \"%s\" finished\n", server.SystemName, server.Mandant)
}
func getMailFromServer(server system, mc *MailCollector) (error){
	baseUrl := fmt.Sprintf("http://%s:%s%s", server.ServerIP, server.Port, server.ServicePath)
	url := fmt.Sprintf("%s/%s?%s&%s", baseUrl, server.FetchSet, "$expand=Attachments", "$format=json")
	if server.Mandant != "" {
		url = fmt.Sprintf("%s&sap-client=%s", url, server.Mandant)
	}
	log.Printf("[URL] GET %s\n", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(server.BasicAuthUser, server.BasicAuthPW)
	req.Header.Add("X-CSRF-Token", "Fetch")

	timeStart := time.Now()
	res, err := client.Do(req)
	mc.setLastFetchDuration(server.SystemName, server.Mandant, time.Since(timeStart))


	if err != nil {
		return fmt.Errorf("[URL] %s", err)
	}
	log.Printf("[URL] Statuscode is %d\n", res.StatusCode)
	if res.StatusCode > 299 {
		return fmt.Errorf("[URL] Statuscode is too high (%d) expected to be below 300", res.StatusCode)
	}

	odata := &odata{}
	json.NewDecoder(res.Body).Decode(odata)
	res.Body.Close()
	for _, v := range odata.Data.Results {
		var logEntries []string
		errorFound := false
		logEntry := fmt.Sprintf("[URL] Found Mail: Objtp=%s, Objyr=%s, Objno=%s with subject %s\n",
		v.Objtp, v.Objyr, v.Objno,v.Subject)
		log.Printf("%s", logEntry)
		logEntries = append(logEntries, logEntry)

		logEntry = fmt.Sprintf("[URL] ContentType is: %s\n", v.MimeType)
		log.Printf("%s", logEntry)
		logEntries = append(logEntries, logEntry)

		for index, attachment := range v.Attachments.Results{
			logEntry = fmt.Sprintf("[URL] Found Attachment on mail, Partnumber is %d\n", attachment.Partno)
			log.Println(logEntry)
			logEntries = append(logEntries,logEntry)

			urlAtt := fmt.Sprintf("%s/%s(Objtp='%s',Objyr='%s',Objno='%s',Partno=%d)/%s",
			baseUrl, server.AttachmentSet,v.Objtp, v.Objyr, v.Objno, attachment.Partno, "$value")

			logEntry = fmt.Sprintf("[URL] GET %s\n", urlAtt)
			log.Printf("%s", logEntry)
			logEntries = append(logEntries, logEntry)

			reqAtt, errAtt := http.NewRequest("GET", urlAtt, nil)
			if errAtt != nil {
				logEntry = fmt.Sprintf("[URL] %s", errAtt)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				logEntry = fmt.Sprintf("[URL] Could not fetch attachment with Partnumber %d\n", attachment.Partno)
				log.Print("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				logEntry = fmt.Sprintf("[URL] ignoring download of attachment Partnumber %d\n", attachment.Partno)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				errorFound = true
				mc.incErrAttachments(server.SystemName, server.Mandant)
				continue
			}
			reqAtt.SetBasicAuth(os.Getenv("BA_USER"), os.Getenv("BA_PASSWORD"))
			reqAtt.Header.Add("X-CSRF-Token", res.Header.Get("X-CSRF-Token"))
			for _, c := range res.Cookies(){
				reqAtt.AddCookie(c)
			}

			logEntry = fmt.Sprintf("[URL] Took Cookies and X-CSRF-Token from originating Request (GET)\n")
			log.Printf("%s", logEntry)
			logEntries = append(logEntries, logEntry)

			resAtt, errAtt := client.Do(reqAtt)
			if errAtt != nil{
				logEntry = fmt.Sprintf("[URL] %s\n", errAtt)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				log.Printf("[URL] Fetch failed for %s\n", urlAtt)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				log.Printf("[URL] ignoring download of attachment Partnumber %d\n", attachment.Partno)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				errorFound = true
				mc.incErrAttachments(server.SystemName, server.Mandant)
				continue
			}
			logEntry = fmt.Sprintf("[URL] Statuscode is %d\n", resAtt.StatusCode)
			log.Printf("%s", logEntry)
			logEntries = append(logEntries, logEntry)
			if resAtt.StatusCode > 299 {
				logEntry = fmt.Sprintf("[URL] Statuscode is too high (%d) expected to be below 300\n", resAtt.StatusCode)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				logEntry = fmt.Sprintf("[URL] ignoring download of attachment Partnumber %d\n", attachment.Partno)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				errorFound = true
				mc.incErrAttachments(server.SystemName, server.Mandant)
				continue
			}

			attachment.data, errAtt = ioutil.ReadAll(resAtt.Body)
			if errAtt != nil{
				logEntry = fmt.Sprintf("[URL] %s\n", errAtt)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				logEntry = fmt.Sprintf("[URL] Could not fetch body\n")
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				logEntry = fmt.Sprintf("[URL] ignoring download of attachment Partnumber %d\n", attachment.Partno)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				errorFound = true
				mc.incErrAttachments(server.SystemName, server.Mandant)
				continue
			}
			logEntry = fmt.Sprintf("[URL] fetched Data in Byte-Format for attachment Partnumber %d\n", attachment.Partno)
			log.Printf("%s", logEntry)
			logEntries = append(logEntries, logEntry)

			attachment.correctLoaded = true
			v.Attachments.Results[index] = attachment

			mc.incAttachments(server.SystemName, server.Mandant)
		}

		logEntry = fmt.Sprintf("[URL] Fetched Mail: Objtp=%s, Objyr=%s, Objno=%s with subject %s\n",
			v.Objtp, v.Objyr, v.Objno, v.Subject)
		log.Printf("%s", logEntry)
		logEntries = append(logEntries, logEntry)

		if errorFound == false {
			err = sendMailSimple(server, v, mc)
			if err != nil {
				logEntry = fmt.Sprintf("%s\n", err)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				errorFound = true
			}
		}

		if errorFound == false {
			urlDelete := fmt.Sprintf("%s/%s", baseUrl, server.FetchSet)
			err = sendDeleteToServer(urlDelete, v, res, client, server)
			if err != nil {
				logEntry = fmt.Sprintf("[URL] %s\n", err)
				log.Printf("%s", logEntry)
				logEntries = append(logEntries, logEntry)

				errorFound = true
			}
		}

		//Notfallmail senden!
		if errorFound == true {
			err = sendMailError(server, logEntries, mc)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil

}


func sendDeleteToServer(uri string, mail mail, resGet *http.Response, client *http.Client, server system) (error){
	url := fmt.Sprintf("%s(Objtp='%s',Objyr='%s',Objno='%s')",
	uri,mail.Objtp, mail.Objyr, mail.Objno, mail.Fortp, mail.Foryr, mail.Forno, mail.Rectp, mail.Recyr, mail.Recno)
	log.Printf("[URL] DELETE %s\n", url)

	reqDel,errDel := http.NewRequest("DELETE", url, nil)
	reqDel.SetBasicAuth(server.BasicAuthUser, server.BasicAuthPW)
	reqDel.Header.Add("X-CSRF-Token", resGet.Header.Get("X-CSRF-Token"))

	for _,c := range resGet.Cookies() {
		reqDel.AddCookie(c)
	}
	log.Println("[URL] Took Cookies and X-CSRF-Token from originating Request (GET)")

	resDel, errDel := client.Do(reqDel)
	if errDel != nil {
		return fmt.Errorf("[URL] %s", errDel)
	}

	log.Printf("[URL] Statuscode is %d\n", resDel.StatusCode)
	if resDel.StatusCode > 299 {
		return fmt.Errorf("[URL] Statuscode is too high (%d) expected to be below 300", resDel.StatusCode)
	}
	log.Printf("[URL] SAP should know we send the mail\n")
	return nil
}
