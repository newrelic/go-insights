package client

import (
	"net/url"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

const (
	testKey = "testKey"
	testID  = "12345"
)

func TestUseCustomURL(t *testing.T) {
	c := &Client{
		Logger: log.New(),
	}

	c.URL, _ = url.Parse(insightsQueryURL)
	assert.Equal(t, c.URL.Scheme, "https", "Schema should be 'https'")

	c.UseCustomURL("http://localhost")
	assert.Equal(t, c.URL.Scheme, "http", "Schema should allow 'http' to be used")
	assert.Equal(t, c.URL.Host, "localhost", "Host should be set")

	c.UseCustomURL("localhost")
	assert.Equal(t, c.URL.Scheme, "https", "Schema should default to 'https'")
}
