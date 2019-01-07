// +build unit

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGZipBuffer_good(t *testing.T) {
	data := []byte{'t', 'e', 's', 't', 0}

	res, err := gZipBuffer(data)

	assert.Nil(t, err, "Error should be nil")
	assert.NotNil(t, res, "Result should not be nil")
}
