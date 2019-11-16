# influx-emoncms
Emoncms server that writes data to influxdb

## Configuration
Run `go run main.go -help` to see configuration options.

## Run the server
`go run main.go`

## Input API

This HTTP server partially implements [the API for writing data to emoncms](https://emoncms.org/site/api#input).

So far, has only been implemented for and tested with an [OpenEVSE](https://www.openevse.com/) station.

Contributions to support more of the API spec and thus more devices is welcome.

### Apikey authentication
If the server is configurated to require an API key, the client must send the correc tkey in the `apikey` field of the request query string or form data.

### Posting Data

Data must be in valid JSON format, contained in the `json` field of the request query string or form data.