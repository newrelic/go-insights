// +build unit

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

/************************************************
 * Tests for the Buffering client
 ************************************************/

// TODO: write unit tests for buffered client
