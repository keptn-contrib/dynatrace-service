package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeframe_ValidArgs(t *testing.T) {
	start := time.Date(2022, 2, 1, 10, 0, 40, 0, time.UTC)
	end := time.Date(2022, 2, 1, 10, 5, 40, 0, time.UTC)

	timeframe, err := NewTimeframe(start, end)
	assert.NoError(t, err)
	assert.EqualValues(t, start, timeframe.Start())
	assert.EqualValues(t, end, timeframe.End())
}

func TestNewTimeframe_InvalidEndBeforeStart(t *testing.T) {
	start := time.Date(2022, 2, 1, 10, 5, 40, 0, time.UTC)
	end := time.Date(2022, 2, 1, 10, 0, 40, 0, time.UTC)

	timeframe, err := NewTimeframe(start, end)
	assert.Error(t, err)
	assert.Nil(t, timeframe)
	assert.Contains(t, err.Error(), "error validating timeframe")
}
