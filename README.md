# sap_mail
This Image is part of a suite, build by Lars Tilsner, Developer at [systema-projekte](https://www.systema-projekte.de)

We found us in the position that our SAP system was firewalled, so it could not communicate itself with a smtp server.
If you have the chance to use SAP Standards, like SCOT/SOST, use them.

Technically this works as a poller. This go program requests an ODATA-Service every 60 Seconds (customizable) and checks for open sendrequests in the.
These are written in SAP table SOSC.

If you need a solution similar to this, we can help you to build an ODATA-Service.
