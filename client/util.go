package client

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
)

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
