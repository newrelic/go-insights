// +build unit

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

/************************************************
 * General insert methods
 ************************************************/
func TestGenerateJSONPostRequest(t *testing.T) {
	var req *http.Request
	var err error

	client := NewInsertClient(testKey, testID)

	// compression: None
	client.Compression = None
	req, err = client.generateJSONPostRequest(testInsertJSON[0])
	assert.NoError(t, err)
	assert.NotNil(t, req)

	// compression: Deflate
	client.Compression = Deflate
	req, err = client.generateJSONPostRequest(testInsertJSON[0])
	assert.NoError(t, err)
	assert.NotNil(t, req)

	// compression: Gzip
	client.Compression = Gzip
	req, err = client.generateJSONPostRequest(testInsertJSON[0])
	assert.NoError(t, err)
	assert.NotNil(t, req)

	// compression: Zlib
	client.Compression = Zlib
	req, err = client.generateJSONPostRequest(testInsertJSON[0])
	assert.NoError(t, err)
	assert.NotNil(t, req)
}

// Successful Insert
func TestJSONPostRequest_success(t *testing.T) {
	var err error

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerSuccess)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.jsonPostRequest(testInsertJSON[0])
	assert.NoError(t, err)
}

// Failed Insert
func TestJSONPostRequest_failure(t *testing.T) {
	var err error

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerFailure)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.jsonPostRequest(testInsertJSON[0])
	assert.Error(t, err)
}

func TestJSONPostRequest_badresponse(t *testing.T) {
	var err error

	// Create a test server to query againt
	ts := httptest.NewServer(testHandlerBad)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.jsonPostRequest(testInsertJSON[0])
	assert.Error(t, err)
}

func TestInsertClientParseResponse(t *testing.T) {
	var err error
	var response *http.Response
	var respGen *httptest.ResponseRecorder

	client := NewInsertClient(testKey, testID)

	// Failure: HTTP Status != 200
	respGen = httptest.NewRecorder()
	respGen.WriteHeader(http.StatusServiceUnavailable)
	_, err = respGen.Write(testInsertResponseJSON["failure"])
	response = respGen.Result()
	err = client.parseResponse(response)
	assert.Error(t, err)

	// Failure: Unable to decode JSON
	respGen = httptest.NewRecorder()
	respGen.WriteHeader(http.StatusOK)
	_, err = respGen.Write(testInsertJSON[0]) // Not a valid JSON response
	response = respGen.Result()
	err = client.parseResponse(response)
	assert.Error(t, err)

	// Success
	respGen = httptest.NewRecorder()
	respGen.WriteHeader(http.StatusOK)
	_, err = respGen.Write(testInsertResponseJSON["success"])
	response = respGen.Result()
	err = client.parseResponse(response)
	assert.NoError(t, err)

}

func TestInsertSendEvents(t *testing.T) {
	var err error

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerSuccess)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.sendEvents(testInsertJSON)
	assert.NoError(t, err)
}

func TestInsertGrabAndConsumeEvents_partial(t *testing.T) {
	var err error

	testData := testInsertJSON

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerSuccess)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.grabAndConsumeEvents(len(testData)-1, testData)
	assert.NoError(t, err)
}

func TestInsertGrabAndConsumeEvents_fullBatch(t *testing.T) {
	var err error

	testData := testInsertJSON

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerSuccess)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	client.BatchSize = len(testData) - 1
	err = client.grabAndConsumeEvents(len(testData)-1, testData)
	assert.NoError(t, err)
}

func TestInsertPostEvent(t *testing.T) {
	var err error

	testData := struct {
		EventType string `json:"eventType"`
		Value     int
		Attribute string
	}{"test", 100, "objectTest"}
	assert.NotNil(t, testData)

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerSuccess)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.PostEvent(testData)
	assert.NoError(t, err)
}

func TestInsertPostEvent_byteArray(t *testing.T) {
	var err error

	testData := []byte(testInsertJSONString)
	assert.NotNil(t, testData)

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerSuccess)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.PostEvent(testData)
	assert.NoError(t, err)
}

func TestInsertPostEvent_string(t *testing.T) {
	var err error

	testData := testInsertJSONString

	// Create a test server to query againt
	ts := httptest.NewServer(testInsertHandlerSuccess)
	defer ts.Close()

	client := NewInsertClient(testKey, testID) // Create test client
	client.URL, err = client.URL.Parse(ts.URL) // Override the URL
	assert.NoError(t, err)
	assert.Equal(t, ts.URL, client.URL.String())

	err = client.PostEvent(testData)
	assert.NoError(t, err)
}

/************************************************
 * Tests for the Non-buffering insert client
 ************************************************/

func TestNewNonBuffer_happy(t *testing.T) {
	c := NewInsertClient(testKey, testID)

	assert.NotNil(t, c)
	assert.Equal(t, testKey, c.InsertKey, "Failed to assign Insert key")
	assert.Contains(t, c.URL.String(), testID, "Failed to generate URL with ID")
	assert.Contains(t, c.URL.String(), "events", "Failed to generate proper insert URL")
	assert.Nil(t, c.eventTimer, "Event timer should be nil")
	assert.Nil(t, c.eventQueue, "Event queue should be nil")
}

func TestValidate_happy(t *testing.T) {
	c := NewInsertClient(testKey, testID)
	validationErr := c.Validate()
	assert.NoError(t, validationErr, "Should not error")
}

func TestValidate_bad(t *testing.T) {
	noKey := NewInsertClient("", testID)
	noKeyErr := noKey.Validate()
	assert.Error(t, noKeyErr, "Empty key should cause error")

	badUrl := NewInsertClient(testKey, "something.org")
	badUrlErr := badUrl.Validate()
	assert.Error(t, badUrlErr, "Non-Newrelic url should cause error")
}

func TestEnqueueNonBuffer_bad(t *testing.T) {
	c := NewInsertClient(testKey, testID)

	if c.eventTimer != nil || c.eventQueue != nil {
		t.Errorf("Unexpected error with new client")
	}

	// There is no queue, so this should fail
	enqueueErr := c.EnqueueEvent("{eventType: \"test\", blah: 1}")

	assert.NotNil(t, enqueueErr)
	assert.Equal(t, "Queueing not enabled for this client", enqueueErr.Error(), "Unknown error returned")

	// Again no queue, so you can't flush it
	flushErr := c.Flush()
	assert.NotNil(t, flushErr)
	assert.Equal(t, "Queueing not enabled for this client", flushErr.Error(), "Unknown error returned")
}

func TestNewInsertClientSetCompression(t *testing.T) {
	client := NewInsertClient(testKey, testID)

	assert.NotNil(t, client)

	client.SetCompression(Gzip)
	assert.Equal(t, Gzip, client.Compression)

	// We only support Gzip right now...
	client.SetCompression(None)
	assert.Equal(t, Gzip, client.Compression)
}

/************************************************
 * Tests for the Buffering client
 ************************************************/

// TODO: write unit tests for buffered client

func TestNewInsertClientEnqueueEvent_good(t *testing.T) {
	client := NewInsertClient(testKey, testID)

	assert.NotNil(t, client)
	client.eventQueue = make(chan []byte, client.BatchSize)

	event := struct {
		Test int
	}{1}
	err := client.EnqueueEvent(event)
	assert.NoError(t, err)
}

func TestNewInsertClientFlush_good(t *testing.T) {
	client := NewInsertClient(testKey, testID)

	assert.NotNil(t, client)
	client.flushQueue = make(chan bool, client.WorkerCount)

	err := client.Flush()
	assert.NoError(t, err)
}
