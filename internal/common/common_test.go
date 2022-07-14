package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampToUnixMillisecondsString(t *testing.T) {
	dt := time.Date(1970, 1, 1, 0, 1, 23, 456, time.UTC)
	expected := "83000" // = (1*60 + 23) * 1000 ms

	got := TimestampToUnixMillisecondsString(dt)

	assert.EqualValues(t, expected, got)
}
