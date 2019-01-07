package client

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Assumption here that responses from insights are either success or error.
type insightsResponse struct {
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success,omitempty"`
}

func (c *InsertClient) jsonPostRequest(body []byte) (err error) {
	const prependText = "Inisghts Post: "

	req, reqErr := c.generateJSONPostRequest(body)
	if reqErr != nil {
		return fmt.Errorf("%s: %v", prependText, reqErr)
	}

	client := &http.Client{Timeout: c.RequestTimeout}
	resp, respErr := client.Do(req)
	if respErr != nil {
		return fmt.Errorf("%s: %v", prependText, respErr)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()

	if parseErr := c.parseResponse(resp); parseErr != nil {
		return fmt.Errorf("%s: %v", prependText, parseErr)
	}

	return nil
}

func (c *InsertClient) generateJSONPostRequest(body []byte) (*http.Request, error) {
	var readBuffer io.Reader
	var buffErr error
	var encoding string

	switch c.Compression {
	case None:
		c.Logger.Debug("Compression: None")
		readBuffer = bytes.NewBuffer(body)
	case Deflate:
		c.Logger.Debug("Compression: Deflate")
		readBuffer = nil
	case Gzip:
		c.Logger.Debug("Compression: Gzip")
		readBuffer, buffErr = gZipBuffer(body)
		encoding = "gzip"
	case Zlib:
		c.Logger.Debug("Compression: Zlib")
		readBuffer = nil
	}

	if buffErr != nil {
		return nil, fmt.Errorf("failed to read body: %v", buffErr)
	}

	request, reqErr := http.NewRequest("POST", c.URL.String(), readBuffer)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to construct request for: %s", body)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Insert-Key", c.InsertKey)
	if encoding != "" {
		request.Header.Add("Content-Encoding", encoding)
	}

	return request, nil
}

func gZipBuffer(body []byte) (io.Reader, error) {
	readBuffer := bufio.NewReader(bytes.NewReader(body))
	buffer := bytes.NewBuffer([]byte{})
	writer := gzip.NewWriter(buffer)

	_, readErr := readBuffer.WriteTo(writer)
	if readErr != nil {
		return nil, readErr
	}

	writeErr := writer.Close()
	if writeErr != nil {
		return nil, writeErr
	}

	return buffer, nil
}

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

	queryResult, err = c.parseQueryResponse(response)
	if err != nil {
		return nil, fmt.Errorf("Failed query: %v", err)
	}

	return queryResult, err
}

func (c *InsertClient) parseResponse(response *http.Response) error {

	body, readErr := ioutil.ReadAll(response.Body)

	if readErr != nil {
		return fmt.Errorf("Failed to read response body: %s", readErr.Error())
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("Bad response from Insights: %d \n\t%s", response.StatusCode, string(body))
	}

	c.Logger.Debugf("Response %d body: %s", response.StatusCode, body)

	respJSON := insightsResponse{}

	if err := json.Unmarshal(body, &respJSON); err != nil {
		return fmt.Errorf("Failed to unmarshal insights response: %v", err)
	}

	// Success
	if response.StatusCode == 200 && respJSON.Success {
		return nil
	}

	// Non 200 response (or 200 not success, if such a thing)
	if respJSON.Error == "" {
		respJSON.Error = "Error unknown"
	}

	return fmt.Errorf("%d: %s", response.StatusCode, respJSON.Error)
}

func (c *QueryClient) generateQueryURL(nrqlQuery string) {
	urlQuery := c.URL.Query()
	urlQuery.Set("nrql", nrqlQuery)
	c.URL.RawQuery = urlQuery.Encode()
	log.Debugf("query url is: %s", c.URL)
}

func (c *QueryClient) parseQueryResponse(response *http.Response) (*QueryResponse, error) {

	body, readErr := ioutil.ReadAll(response.Body)

	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %s", readErr.Error())
	}

	c.Logger.Debugf("Response %d body: %s", response.StatusCode, body)

	parsedResponse := &QueryResponse{}

	if jsonErr := json.Unmarshal(body, parsedResponse); jsonErr != nil {
		return nil, fmt.Errorf("Unable to unmarshal query response: %v", jsonErr)
	}

	return parsedResponse, nil
}
