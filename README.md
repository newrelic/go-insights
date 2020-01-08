# New Relic Insights Client
A Go library for interacting with insights.

[![CircleCI](https://circleci.com/gh/newrelic/go-insights.svg?style=svg)](https://circleci.com/gh/newrelic/go-insights)
[![Go Report Card](https://goreportcard.com/badge/github.com/newrelic/go-insights?style=flat-square)](https://goreportcard.com/report/github.com/newrelic/go-insights)
[![GoDoc](https://godoc.org/github.com/newrelic/go-insights?status.svg)](https://godoc.org/github.com/newrelic/go-insights)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/newrelic/go-insights/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/newrelic/go-insights?style=flat-square)](https://github.com/newrelic/go-insights/releases/latest)


## Disclaimer
New Relic has open-sourced this integration to enable monitoring of this technology. This integration is provided AS-IS WITHOUT WARRANTY OR SUPPORT, although you can report issues and contribute to this integration via GitHub. Support for this integration is available with an [Expert Services subscription](newrelic.com/expertservices).

## Contents

There are two parts to this library:
* The [Command Line Interface](#cli) is a compiled standalone binary that can be executed to insert or query Newrelic Insights.
* The [Insights Client Library](#insights-client-library) can be used in a go project for posting to and querying events from insights.

## CLI

The Insights CLI can be used to query and insert data in insights.
The following commands and flags are supported:

```
usage: go-insights --key=KEY --id=ID [<flags>] <command> [<args> ...]

Flags:
  --help         Show help (also see --help-long and --help-man).
  -k, --key=KEY  Your insights key.
  -i, --id=ID    Your New Relic account ID
  -u, --url=URL  Custom insights endpoint.
  --version      Show application version.
  -d, --debug    Enable debug level logging.

Commands:
  help [<command>...]
    Show help.

  insert <file path>
    Insert data to insights.

  query <query string>
    Query data in insights.
```

## Insights Client Library
The client library has two functions. It contains a query client and an insert client.

### Query Client
The query client will make an API call to insights and return the results of your query in a QueryResponse struct:

```go
type QueryResponse struct {
  Results  []map[string]interface{} `json:"results"`
  Metadata QueryMetadata            `json:"metadata"`
}
```

### Insert Client
The insert client will insert data into insights.
There are two methods of use. You can send single events one at a time. Alternatively, you can run the client in batch mode, which runs a goroutine and sends
events to insights in batches.

#### Sending Single Events
```go
package main

import (
  "fmt"

  insights "github.com/newrelic/go-insights/client"
)

type TestType struct {
  EventType    string `json:"eventType"`
  AwesomeScore int    `json:"AwesomeScore"`
}

insightAccountID := "0"
insightInsertKey := "abc123example"

client := insights.NewInsertClient(insightInsertKey, insightAccountID)
if validationErr := client.Validate(); validationErr != nil {
  //however it is appropriate to handle this in your use case
  log.Errorf("Validation Error!")
}

testData := TestType{
  EventType:    "testEvent",
  AwesomeScore: 25,
}

if postErr := client.PostEvent(testData); postErr != nil {
  log.Errorf("Error: %v\n", err)
}
```

#### Enqueueing Events in Batch Mode

```go
package main

import (
  "fmt"
  "time"

  insights "github.com/newrelic/go-insights/client"
)

type TestType struct {
  EventType    string `json:"eventType"`
  AwesomeScore int    `json:"AwesomeScore"`
}

func main() {
  insightAccountID := "0"
  insightInsertKey := "abc123example"

  // Create the client instance
  client := insights.NewInsertClient(insightInsertKey, insightAccountID)
  if validationErr := client.Validate(); validationErr != nil {
    //however it is appropriate to handle this in your use case
    log.Errorf("Validation Error!")
  }
  if startError := client.Start(); startError != nil {
    log.Errorf("failed to start client")
  }

  // Add some data with delay
  for x := 0; x < 10; x++ {
    fmt.Printf("Enqueueing data: %d\n", x)
    err := client.EnqueueEvent(TestType{EventType: "YourTable", AwesomeScore: 9000 + x})
    if err != nil {
      fmt.Printf("Error: %v\n", err)
      return
    }
    time.Sleep(10 * time.Second)
  }

  // Make sure nothing is pending in the queue
  client.Flush()
}
```
