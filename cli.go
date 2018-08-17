package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/newrelic/go-insights/client"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.0.0"

var (
	insightsKey = kingpin.Flag("key", "Your insights key.").Short('k').Required().String()
	accountID   = kingpin.Flag("id", "Your New Relic account ID").Short('i').Required().String()
	insightsURL = kingpin.Flag("url", "Custom insights endpoint.").Short('u').String()

	insertCmd = kingpin.Command("insert", "Insert data to insights.")
	dataFile  = insertCmd.Arg("file path", "Path to file containing data to insert.").Required().String()

	queryCmd    = kingpin.Command("query", "Query data in insights.")
	queryString = queryCmd.Arg("query string", "Insights Query").Required().String()

	versionFlag = kingpin.Version("Insights client version: " + version)
	logDebug    = kingpin.Flag("debug", "Enable debug level logging.").Short('d').Bool()
)

func main() {
	cmd := kingpin.Parse()

	if *logDebug {
		log.SetLevel(log.DebugLevel)
		log.Debug("debug logging is on")
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Debugf("Creating new %s client", cmd)

	switch cmd {
	case "insert":
		cli := client.NewInsertClient(*insightsKey, *accountID)
		if cli == nil {
			log.Fatal("Failed to create a %s client", cmd)
		}
		if len(*insightsURL) > 0 {
			cli.UseCustomURL(*insightsURL)
		}

		if err := cli.Validate(); err != nil {
			log.Fatalf("Insert Client configuration validation failed: %s", err.Error())
		}
		data, fileErr := ioutil.ReadFile(*dataFile)
		if fileErr != nil {
			log.Errorf("Error reading event data file: %v", fileErr)
		}

		if postErr := cli.PostEvent(data); postErr != nil {
			log.Errorf("Insert error: %v", postErr)
		}

	case "query":
		cli := client.NewQueryClient(*insightsKey, *accountID)
		if len(*insightsURL) > 0 {
			cli.UseCustomURL(*insightsURL)
		}

		if err := cli.Validate(); err != nil {
			log.Fatalf("Insert Client configuration validation failed: %s", err.Error())
		}
		result, queryErr := cli.QueryEvents(*queryString)
		if queryErr != nil {
			log.Fatal(queryErr)
		}

		resultJSON, jsonErr := json.MarshalIndent(result.Results, "", "  ")
		if jsonErr != nil {
			log.Errorf("Unable to JSON format reulst data: %s", result.Results)
		}

		fmt.Printf("%s\n", resultJSON)
	default:
		log.Fatal("Unknown command")
	}

	return
}
