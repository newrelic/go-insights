// +build unit

package client

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQueryClient(t *testing.T) {
	client := NewQueryClient(testKey, testID)
	clientURL := createQueryURL(testID)

	// In order as created in NewQueryClient
	assert.NotNil(t, client, "NewQueryClient response is nil")
	assert.IsType(t, &QueryClient{}, client, "NewQueryClient should return a *QueryClient")
	assert.Equal(t, clientURL, client.URL, "NewQueryClient should of set URL")
	assert.Equal(t, testKey, client.QueryKey, "NewQueryClient should of set QueryKey")
	assert.NotNil(t, client.Logger, "NewQueryClient should of set Logger")
	assert.Equal(t, DefaultQueryRequestTimeout, client.RequestTimeout, "NewQueryClient should of set a default request timeout")
	assert.Equal(t, DefaultRetries, client.RetryCount, "NewQueryClient should of set a default number of request retries")
	assert.Equal(t, DefaultRetryWaitTime, client.RetryWait, "NewQueryClient should of set a default retry wait time")
}

func TestQueryClientValidate(t *testing.T) {
	var err error

	client := NewQueryClient(testKey, testID)

	assert.NotNil(t, client, "NewQueryClient response is nil")

	err = client.Validate()
	assert.NoError(t, err, "Client should successfully Validate")

	// Test URL validation
	client.URL = createQueryURL("ASDF")
	err = client.Validate()
	assert.Error(t, err, "client validation should of failed with invalid URL")
	client.URL = createQueryURL(testID)

	client.QueryKey = ""
	err = client.Validate()
	assert.Error(t, err, "client validation should of failed with invalid query key")
}

func TestGenerateQueryURL(t *testing.T) {
	client := NewQueryClient(testKey, testID)
	testURL := client.URL.String() + "?" + testNRQLQueryEncoded
	query, err := client.generateQueryURL(testNRQLQuery)
	assert.NoError(t, err)
	assert.Equal(t, testURL, query, "generateQueryURL did not properly encode URL")
}

func TestQueryClientQuery(t *testing.T) {
	var err error
	res := &QueryResponse{}

	// Create a test server to query againt
	ts := httptest.NewServer(testQueryHandlerEmpty)
	defer ts.Close()

	client := NewQueryClient(testKey, testID)  // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	// Empty NRQL
	err = client.Query("", res)
	assert.Error(t, err, "Empty NRQL query should fail")

	// Nil interface{}
	err = client.Query(testNRQLQuery, nil)
	assert.Error(t, err, "Empty NRQL query should fail")

	// valid case
	err = client.Query(testNRQLQuery, res)
	assert.NoError(t, err)

	// Custom struct
	cust := struct {
		Results  []map[string][]string `json:"results"`
		Metadata QueryMetadata         `json:"metadata"`
	}{}
	err = client.Query(testNRQLQuery, &cust)
	assert.NoError(t, err)
}

func TestQueryClientQueryRequest(t *testing.T) {
	var err error
	var res *QueryResponse

	// Create a test server to query againt
	ts := httptest.NewServer(testHandlerBad)
	defer ts.Close()

	client := NewQueryClient(testKey, testID)  // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	// Empty NRQL
	res = &QueryResponse{}
	query := ""
	err = client.queryRequest(query, res)
	assert.Error(t, err, "Empty NRQL query should fail")
}

func TestQueryClientQueryRequest_decodeFailure(t *testing.T) {
	var err error

	// Create a test server to query againt
	ts := httptest.NewServer(testQueryHandlerEmpty)
	defer ts.Close()

	client := NewQueryClient(testKey, testID)  // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	// NIL result pointer
	query, err := client.generateQueryURL(testNRQLQuery)
	assert.NoError(t, err)
	err = client.queryRequest(query, nil)
	assert.Error(t, err, "Empty result pointer should fail")
}

func TestQueryClientQueryEvents_bad(t *testing.T) {
	var err error
	var resp *QueryResponse

	// Create a test server to query againt
	ts := httptest.NewServer(testHandlerBad)
	defer ts.Close()

	client := NewQueryClient(testKey, testID)  // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	// Invalid requests
	resp, err = client.QueryEvents("")
	assert.Error(t, err, "Empty NRQL query should fail")
	assert.Nil(t, resp, "Response should be nil on failed query")

	// Valid requests with Server Errors
	resp, err = client.QueryEvents(testNRQLQuery)
	assert.Error(t, err, "Server Errors should be returned to caller")
	assert.Nil(t, resp, "Response should be nil on server error")

	// Valid requests with Invalid Server
	client.URL, err = client.URL.Parse("http://localhost:0")
	assert.NoError(t, err)
	resp, err = client.QueryEvents(testNRQLQuery)
	assert.Error(t, err, "Server Errors should be returned to caller")
	assert.Nil(t, resp, "Response should be nil on server error")
}

func TestQueryClientQueryEvents_good(t *testing.T) {
	var err error
	var resp *QueryResponse

	// Create a test server to query againt
	ts := httptest.NewServer(testQueryHandlerEmpty)
	defer ts.Close()

	client := NewQueryClient(testKey, testID)  // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	resp, err = client.QueryEvents(testNRQLQuery)
	assert.NoError(t, err, "Valid query to test server should not return error")
	assert.NotNil(t, resp, "Response should not be nil")
}
