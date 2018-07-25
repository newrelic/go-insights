package client

import (
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	insightsInsertURL = "https://insights-collector.newrelic.com/v1/accounts"
	insightsQueryURL  = "https://insights-api.newrelic.com/v1/accounts"
)

// Compression to use during transport.
type Compression int32

// Supported / recognized types of compression
const (
	None    Compression = iota
	Deflate Compression = iota
	Gzip    Compression = iota
	Zlib    Compression = iota
)

// Client is the building block of the insert and query clients
type Client struct {
	URL            *url.URL
	Logger         *log.Logger
	RequestTimeout time.Duration
	RetryCount     int
	RetryWait      time.Duration
}

// UseCustomURL allows overriding the default Insights Host / Scheme.
func (c *Client) UseCustomURL(customURL string) {
	newURL, _ := url.Parse(customURL)
	if len(newURL.Scheme) < 1 {
		c.URL.Scheme = "https"
	} else {
		c.URL.Scheme = newURL.Scheme
	}

	c.URL.Host = newURL.Host
	c.Logger.Debugf("Using custom URL: %s", c.URL)
}
