package client

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

// QueryClient contains all of the configuration required for queries
type QueryClient struct {
	QueryKey string
	Client
}

// QueryResponse used to decode the JSON response from Insights
type QueryResponse struct {
	Results  []map[string]interface{} `json:"results"`
	Metadata QueryMetadata            `json:"metadata"`
}

// QueryMetadata used to decode the JSON response metadata from Insights
type QueryMetadata struct {
	Contents        []interface{} `json:"contents"`
	EventType       string        `json:"eventType"`
	OpenEnded       bool          `json:"openEnded"`
	BeginTime       time.Time     `json:"beginTime"`
	EndTime         time.Time     `json:"endTime"`
	BeginTimeMillis int64         `json:"beginTimeMillis"`
	EndTimeMillis   int64         `json:"endTimeMillis"`
	RawSince        string        `json:"rawSince"`
	RawUntil        string        `json:"rawUntil"`
	RawCompareWith  string        `json:"rawCompareWith"`
}

// NewQueryClient makes a new client for the user to query with.
func NewQueryClient(queryKey, accountID string) *QueryClient {
	client := &QueryClient{}
	client.URL = createQueryURL(accountID)
	client.QueryKey = queryKey
	client.Logger = log.New()

	// Defaults
	client.RequestTimeout = 10 * time.Second
	client.RetryCount = 3
	client.RetryWait = 5 * time.Second

	return client
}

func createQueryURL(accountID string) *url.URL {
	insightsURL, _ := url.Parse(insightsQueryURL)
	insightsURL.Path = fmt.Sprintf("%s/%s/query", insightsURL.Path, accountID)
	return insightsURL
}

// Validate makes sure the QueryClient is configured correctly for use
func (c *QueryClient) Validate() error {
	if correct, _ := regexp.MatchString("api.newrelic.com/v1/accounts/[0-9]+/query", c.URL.String()); !correct {
		return fmt.Errorf("Invalid query endpoint %s", c.URL)
	}

	if len(c.QueryKey) < 1 {
		return fmt.Errorf("Not a valid license key: %s", c.QueryKey)
	}
	return nil
}

// QueryEvents initiates an Insights query, returns a response for parsing
func (c *QueryClient) QueryEvents(nrqlQuery string) (*QueryResponse, error) {
	c.Logger.Debugf("Querying: %s", nrqlQuery)
	c.generateQueryURL(nrqlQuery)

	response, queryErr := c.queryRequest(nrqlQuery)
	if queryErr != nil {
		return nil, queryErr
	}

	return response, nil
}
