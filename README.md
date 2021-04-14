# sap_mail
This Image is part of a suite, build by Lars Tilsner, Developer at [systema-projekte](https://www.systema-projekte.de)

We found us in the position that our SAP system was firewalled, so it could not communicate itself with a smtp server.
If you have the chance to use SAP Standards, like SCOT/SOST, use them.

Technically this works as a poller. This go program requests an ODATA-Service every 60 Seconds (customizable) and checks for open sendrequests in the.
These are written in SAP table SOSC.

If you need a solution similar to this, we can help you to build an ODATA-Service.

## Environment Variables

### SMTP_HOST

Server name of your smtp server, it can be a hostname or an ip.

### SMTP_PORT

Port to use for SMTP connection. It is preferred to use 465, which is a ssl port.

### SMTP_USER

The user to identify against the smtp server.

### SMTP_PASSWORD

The passwort to identify against the smtp server.

### SMTP_FROM

The SMTP Mail from which the mails should be send.

### GATEWAY

The Gateway to reach your SAP-system.
This is probably your VPN-Gateway.

### ERROR_MAIL

The Mail Recipient in case of every error.
In case of errors, the belonging logs will be sent to this Recipient

### ERROR_SUBJECT

The prefix of the error mail. 
It will be enhanced at runtime with the system and mandant, which caused the error.
The Subject will look like ERROR_SUBJECT SID[MANDT]

### MAIL_OFF

You can disable the mails, if you want to test the integration with the SOST.

## Prometheus

This image provides basic statistics via Prometheus. Just scrape it at port 8080.

