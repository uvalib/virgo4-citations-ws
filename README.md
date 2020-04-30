# Virgo4 Citations Web Service

This is a web service to generate citations from Virgo4 records.

* GET /version : returns build version
* GET /healthcheck : returns health check information
* GET /metrics : returns Prometheus metrics
* GET /api/ris/?item={url} : generates a RIS file from the record returned by url.

All endpoints under /api require authentication.

### System Requirements

* GO version 1.12.0 or greater
