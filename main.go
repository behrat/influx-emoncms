package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var (
	// HTTP Server Config
	apiKey     = flag.String("apikey", "", "API key that client must use (Authentiation disabled if not given)")
	listenAddr = flag.String("listen", ":8080", "Address for HTTP server to listen on")

	// Database Config
	dbAddr          = flag.String("db", "http://127.0.0.1:8086", "Influxdb server address")
	dbName          = flag.String("db-name", "emoncms", "Database name")
	measurementName = flag.String("measurement", "emoncms", "Measurement name")
)

func handleError(w http.ResponseWriter, r *http.Request, statusCode int, reason string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(reason))
	w.Write([]byte("\n"))
	log.Printf("Got bad rquest from %s: %s", r.RemoteAddr, reason)
}

func main() {
	flag.Parse()

	dbClient, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr: *dbAddr,
	})
	if err != nil {
		log.Fatal("Error creating InfluxDB Client: ", err.Error())
	}
	defer dbClient.Close()

	if _, err := dbClient.Query(
		influxdb.NewQuery(
			fmt.Sprintf("CREATE DATABASE %s", *dbName),
			"", ""),
	); err != nil {
		log.Fatal("Error creating database: ", err)
	}

	http.HandleFunc("/input/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		node := r.Form.Get("node")
		if node == "" {
			handleError(w, r, http.StatusBadRequest, "No node given")
			return
		}

		if *apiKey != "" {
			reqApiKey := r.Form.Get("apikey")
			if reqApiKey == "" {
				handleError(w, r, http.StatusBadRequest, "No apikey given")
				return
			}
			if reqApiKey != *apiKey {
				handleError(w, r, http.StatusUnauthorized, "Wrong apikey")
				return
			}
		}

		jsonData := make(map[string]interface{})
		jsonString := r.Form.Get("json")
		err := json.Unmarshal([]byte(jsonString), &jsonData)
		if err != nil {
			handleError(w, r, http.StatusBadRequest, fmt.Sprint("Error unmarshalling json: ", err))
			return
		}

		bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
			Database:  *dbName,
			Precision: "s",
		})
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, fmt.Sprint("Error creating BatchPoints: ", err))
			return
		}

		point, err := influxdb.NewPoint(
			*measurementName,
			map[string]string{
				"node": node,
			},
			jsonData,
			time.Now(),
		)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, fmt.Sprint("Error creating Point: ", err))
			return
		}

		bp.AddPoint(point)

		err = dbClient.Write(bp)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, fmt.Sprint("Error writing to databse: ", err))
			return
		}
	})

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
