# Virgo4 Citations Web Service

This is a web service to generate citations from Virgo4 records.

* GET /version : returns build version
* GET /healthcheck : returns health check information
* GET /format/ris?item={url} : generates a RIS file from the V4 record returned by url

### System Requirements

* GO version 1.26.10 or greater
