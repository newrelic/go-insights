// +build unit

package client

import (
	"net/http"
)

// Fixture stuff
var testNRQLQuery string = "SHOW eventtypes"
var testNRQLQueryEncoded string = "nrql=SHOW+eventtypes"
var testNRQLResponseJSON []byte = []byte(`{"results": [{"eventTypes": [] }], "metadata": {"guid": "d87afea4-083f-29c9-3390-8ab07e271455", "routerGuid": "", "messages": [], "contents": [{"function": "eventTypes"}]}}`)

var testInsertJSON [][]byte = [][]byte{
	[]byte(`{"eventType": "test", "num": 0, "str": "test" }`),
	[]byte(`{"eventType": "test", "num": 1, "str": "test" }`),
	[]byte(`{"eventType": "test", "num": 2, "str": "test" }`),
	[]byte(`{"eventType": "test", "num": 3, "str": "test" }`),
}
var testInsertJSONString string = `{"eventType": "test", "num": 0, "str": "test" }`

var testInsertJSONBad []byte = []byte(`"num": 1, "str": `)
var testInsertResponseJSON map[string][]byte = map[string][]byte{
	"success": []byte(`{"success": true, "error": "" }`),
	"failure": []byte(`{"success": false, "error": "some random error"}`),
}

var testHandlerBad http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) })

var testQueryHandlerEmpty http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(testNRQLResponseJSON)
})

var testInsertHandlerSuccess http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(testInsertResponseJSON["success"])
})

var testInsertHandlerFailure http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(testInsertResponseJSON["failure"])
})
