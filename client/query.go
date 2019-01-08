package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// NewQueryClient makes a new client for the user to query with.
func NewQueryClient(queryKey, accountID string) *QueryClient {
	client := &QueryClient{}
	client.URL = createQueryURL(accountID)
	client.QueryKey = queryKey
	client.Logger = log.New()

	// Defaults
	client.RequestTimeout = DefaultQueryRequestTimeout
	client.RetryCount = DefaultRetries
	client.RetryWait = DefaultRetryWaitTime

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
func (c *QueryClient) QueryEvents(nrqlQuery string) (response *QueryResponse, err error) {
	c.Logger.Debugf("Querying: %s", nrqlQuery)
	err = c.generateQueryURL(nrqlQuery)
	if err != nil {
		return nil, err
	}

	response = &QueryResponse{}
	err = c.queryRequest(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// queryRequest makes a NRQL query
func (c *QueryClient) queryRequest(queryResult interface{}) (err error) {
	var request *http.Request
	var response *http.Response

	if len(c.URL.RawQuery) < 1 {
		return fmt.Errorf("Query string can not be empty")
	}

	if queryResult == nil {
		return errors.New("Must have pointer for result")
	}

	request, err = http.NewRequest("GET", c.URL.String(), nil)
	if err != nil {
		return err
	}

	request.Header.Add("Accept", "application/json")
	request.Header.Add("X-Query-Key", c.QueryKey)

	client := &http.Client{Timeout: c.RequestTimeout}

	response, err = client.Do(request)
	if err != nil {
		err = fmt.Errorf("Failed query request for: %v", err)
		return
	}
	defer func() {
		respErr := response.Body.Close()
		if respErr != nil && err == nil {
			err = respErr // Don't mask previous errors
		}
	}()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("Bad response code: %d", response.StatusCode)
		return
	}

	err = c.parseResponse(response, queryResult)
	if err != nil {
		err = fmt.Errorf("Failed query: %v", err)
	}

	return err
}

// generateQueryURL URL encodes the NRQL
func (c *QueryClient) generateQueryURL(nrqlQuery string) error {
	if len(nrqlQuery) < 10 {
		return fmt.Errorf("Invalid query [%s]", nrqlQuery)
	}

	urlQuery := c.URL.Query()
	urlQuery.Set("nrql", nrqlQuery)
	c.URL.RawQuery = urlQuery.Encode()

	log.Debugf("query url is: %s", c.URL)

	return nil
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
