package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// InsertClient contains all of the configuration required for inserts
type InsertClient struct {
	InsertKey   string
	eventQueue  chan []byte
	eventTimer  *time.Timer
	flushQueue  chan bool
	WorkerCount int
	BatchSize   int
	BatchTime   time.Duration
	Compression Compression
	Client
	Statistics
}

// Statistics about the inserted data
type Statistics struct {
	EventCount         int64
	FlushCount         int64
	ByteCount          int64
	FullFlushCount     int64
	PartialFlushCount  int64
	TimerExpiredCount  int64
	InsightsRetryCount int64
	HTTPErrorCount     int64
}

// NewInsertClient makes a new client for the user to send data with
func NewInsertClient(insertKey string, accountID string) *InsertClient {
	client := &InsertClient{}
	client.URL = createInsertURL(accountID)
	client.InsertKey = insertKey
	client.Logger = log.New()
	client.Compression = None

	// Defaults
	client.RequestTimeout = 10 * time.Second
	client.RetryCount = 3
	client.RetryWait = 5 * time.Second

	// Defaults for buffered client.
	// These are here so they can be overwritten before calling start().
	client.WorkerCount = 1
	client.BatchTime = 1 * time.Minute
	client.BatchSize = 950

	return client
}

func createInsertURL(accountID string) *url.URL {
	insightsURL, _ := url.Parse(insightsInsertURL)
	insightsURL.Path = fmt.Sprintf("%s/%s/events", insightsURL.Path, accountID)
	return insightsURL
}

// Start runs the insert client in batch mode.
func (c *InsertClient) Start() error {
	if c.eventQueue != nil {
		return errors.New("Insights client already in daemon mode")
	}

	c.eventQueue = make(chan []byte, c.BatchSize)
	c.eventTimer = time.NewTimer(c.BatchTime)
	c.flushQueue = make(chan bool, c.WorkerCount)

	// TODO: errors returned from the call to watchdog()
	// and batchWorker() are simply dropped on the floor.
	go c.watchdog()
	go c.batchWorker()
	c.Logger.Infof("Insights client launched in daemon mode with endpoint %s", c.URL)

	return nil
}

// StartListener creates a goroutine that consumes from a channel and
// Enqueues events as to not block the writing of events to the channel
//
func (c *InsertClient) StartListener(inputChannel chan interface{}) (err error) {
	// Allow this to be called instead of Start()
	if c.eventQueue == nil {
		if err = c.Start(); err != nil {
			return err
		}
	}
	if inputChannel == nil {
		return errors.New("Channel to listen is nil")
	}

	go c.queueWorker(inputChannel)

	c.Logger.Info("Insights client started channel listener")

	return nil
}

// Validate makes sure the InsertClient is configured correctly for use
func (c *InsertClient) Validate() error {
	if correct, _ := regexp.MatchString("collector.newrelic.com/v1/accounts/[0-9]+/events", c.URL.String()); !correct {
		return fmt.Errorf("Invalid insert endpoint %s", c.URL)
	}

	if len(c.InsertKey) < 1 {
		return fmt.Errorf("Not a valid license key: %s", c.InsertKey)
	}
	return nil
}

// EnqueueEvent handles the queueing. Only works in batch mode.
func (c *InsertClient) EnqueueEvent(data interface{}) (err error) {
	if c.eventQueue == nil {
		return errors.New("Queueing not enabled for this client")
	}

	var jsonData []byte
	atomic.AddInt64(&c.Statistics.EventCount, 1)

	if jsonData, err = json.Marshal(data); err != nil {
		return err
	}

	c.eventQueue <- jsonData

	return err
}

// PostEvent allows sending a single event directly.
func (c *InsertClient) PostEvent(data interface{}) error {
	var jsonData []byte

	switch data.(type) {
	case []byte:
		jsonData = data.([]byte)
	case string:
		jsonData = []byte(data.([]byte))
	default:
		var jsonErr error
		jsonData, jsonErr = json.Marshal(data)
		if jsonErr != nil {
			return fmt.Errorf("Error marshaling event data: %s", jsonErr.Error())
		}
	}

	// Needs to handle array of events. maybe pull into separate validation func
	if !strings.Contains(string(jsonData), "eventType") {
		return fmt.Errorf("Event data must contain eventType field. %s", jsonData)
	}

	c.Logger.Debugf("Posting to insights: %s", jsonData)

	if requestErr := c.jsonPostRequest(jsonData); requestErr != nil {
		return requestErr
	}

	return nil
}

// Flush gives the user a way to manually flush the queue in the foreground.
// This is also used by watchdog when the timer expires.
func (c *InsertClient) Flush() error {
	c.Logger.Debug("Flushing insights client")
	if c.flushQueue == nil {
		return errors.New("Queueing not enabled for this client")
	}
	atomic.AddInt64(&c.Statistics.FlushCount, 1)

	c.flushQueue <- true

	return nil
}

//
// queueWorker watches a channel and Enqueues items as they appear so
// we don't block on EnqueueEvent
//
func (c *InsertClient) queueWorker(inputChannel chan interface{}) (err error) {
	for {
		select {
		case msg := <-inputChannel:
			c.EnqueueEvent(msg)
		}
	}
}

//
// watchdog has a Timer that will send the results once the
// it has expired.
//
func (c *InsertClient) watchdog() (err error) {
	for {
		select {
		case <-c.eventTimer.C:
			// Timer expired, and we have data, send it
			atomic.AddInt64(&c.Statistics.TimerExpiredCount, 1)
			c.Logger.Debug("Timeout expired, flushing queued events")
			if flushErr := c.Flush(); flushErr != nil {
				c.Logger.Errorf("Flush error: %s", flushErr.Error())
			}
			c.eventTimer.Reset(c.BatchTime)
		}
	}
}

//
// batchWorker reads []byte from the queue until a threshold is passed,
// then copies the []byte it has read and sends that batch along to Insights
// in its own goroutine.
//
func (c *InsertClient) batchWorker() (err error) {
	eventBuf := make([][]byte, c.BatchSize, c.BatchSize)
	count := 0
	for {
		select {
		case item := <-c.eventQueue:
			eventBuf[count] = item
			count++
			if count >= c.BatchSize {
				err = c.grabAndConsumeEvents(count, eventBuf)
				count = 0
			}
		case <-c.flushQueue:
			if count > 0 {
				err = c.grabAndConsumeEvents(count, eventBuf)
				count = 0
			}
		}
	}
}

// grabAndConsumeEvents makes a copy of the event handles,
// and asynchronously writes those events in its own goroutine.
// The write is attempted up to c.RetryCount times.
//
// TODO: Any errors encountered doing the write are dropped on the floor.
// Even the last error (in the event of trying c.RetryCount times)
// is dropped.
//
func (c *InsertClient) grabAndConsumeEvents(count int, eventBuf [][]byte) (err error) {
	if count < c.BatchSize-20 {
		atomic.AddInt64(&c.Statistics.PartialFlushCount, 1) // Allow for some fuzz, although there should be none
	} else {
		atomic.AddInt64(&c.Statistics.FullFlushCount, 1)
	}

	saved := make([][]byte, count, count)
	for i := 0; i < count; i++ {
		saved[i] = eventBuf[i]
		eventBuf[i] = nil
	}

	go func(count int, saved [][]byte) {
		// only send the slice that we pulled into the buffer
		for tries := 0; tries < c.RetryCount; tries++ {
			if sendErr := c.sendEvents(saved[0:count]); sendErr != nil {
				c.Logger.Errorf("Failed to send events [%d/%d]: %v", tries, c.RetryCount, sendErr)
				atomic.AddInt64(&c.Statistics.InsightsRetryCount, 1)
				time.Sleep(c.RetryWait)
			} else {
				break
			}
		}
	}(count, saved)

	return nil
}

// sendEvents accepts a slice of marshalled JSON and sends it to Insights
//
func (c *InsertClient) sendEvents(events [][]byte) error {
	var buf bytes.Buffer

	// Since we already marshalled all of the data into JSON, let's make a
	// hand-crafted, artisanal JSON array
	buf.WriteString("[")
	eventCount := len(events) - 1
	for e := range events {
		buf.Write(events[e])
		if e < eventCount {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	atomic.AddInt64(&c.Statistics.ByteCount, int64(buf.Len()))

	if c.URL == nil {
		// TODO: Somewhat of a hack for the test suite, should mock this
		return nil
	}

	if postErr := c.jsonPostRequest(buf.Bytes()); postErr != nil {
		return postErr
	}

	return nil
}

// SetCompression allows modification of the compression type used in communication
//
func (c *InsertClient) SetCompression(compression Compression) {
	c.Compression = Gzip
	// use gzip only for now
	// c.Compression = compression
	log.Debugf("Compression set: %d", c.Compression)
}
