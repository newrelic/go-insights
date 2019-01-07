package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	Facets   []map[string]interface{} `json:"facets"`
	Metadata QueryMetadata            `json:"metadata"`
}

// QueryMetadata used to decode the JSON response metadata from Insights
type QueryMetadata struct {
	Contents        interface{} `json:"contents"`
	EventType       string      `json:"eventType"`
	OpenEnded       bool        `json:"openEnded"`
	BeginTime       time.Time   `json:"beginTime"`
	EndTime         time.Time   `json:"endTime"`
	BeginTimeMillis int64       `json:"beginTimeMillis"`
	EndTimeMillis   int64       `json:"endTimeMillis"`
	RawSince        string      `json:"rawSince"`
	RawUntil        string      `json:"rawUntil"`
	RawCompareWith  string      `json:"rawCompareWith"`
}

// NewQueryClient makes a new client for the user to query with.
func NewQueryClient(queryKey, accountID string) *QueryClient {
	client := &QueryClient{}
	client.URL = createQueryURL(accountID)
	client.QueryKey = queryKey
	client.Logger = log.New()

	// Defaults
	client.RequestTimeout = 20 * time.Second
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

// queryRequest makes a NRQL query
func (c *QueryClient) queryRequest(nrqlQuery string) (queryResult *QueryResponse, err error) {
	if len(c.URL.RawQuery) < 1 {
		return nil, fmt.Errorf("Query string can not be empty")
	}

	request, reqErr := http.NewRequest("GET", c.URL.String(), nil)
	if reqErr != nil {
		return nil, fmt.Errorf("Failed to construct request for: %s", nrqlQuery)
	}

	request.Header.Add("Accept", "application/json")
	request.Header.Add("X-Query-Key", c.QueryKey)

	client := &http.Client{Timeout: c.RequestTimeout}

	response, respErr := client.Do(request)

	if respErr != nil {
		return nil, fmt.Errorf("Failed query request for: %v", respErr)
	}

	defer func() {
		err = response.Body.Close()
	}()

	err = c.parseResponse(response, queryResult)
	if err != nil {
		return nil, fmt.Errorf("Failed query: %v", err)
	}

	return queryResult, err
}

// generateQueryURL URL encodes the NRQL
func (c *QueryClient) generateQueryURL(nrqlQuery string) {
	urlQuery := c.URL.Query()
	urlQuery.Set("nrql", nrqlQuery)
	c.URL.RawQuery = urlQuery.Encode()
	log.Debugf("query url is: %s", c.URL)
}

// parseQueryResponse takes an HTTP response, make sure it is a valid response,
// then attempts to decode the JSON body into the `parsedResponse` interface
func (c *QueryClient) parseResponse(response *http.Response, parsedResponse interface{}) error {
	body, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return fmt.Errorf("failed to read response body: %s", readErr.Error())
	}

	c.Logger.Debugf("Response %d body: %s", response.StatusCode, body)

	if jsonErr := json.Unmarshal(body, parsedResponse); jsonErr != nil {
		return fmt.Errorf("Unable to unmarshal query response: %v", jsonErr)
	}

	return nil
}
