package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeframeParser_ValidArgs(t *testing.T) {
	expectedTimeframe, err := NewTimeframe(time.Date(2022, 2, 1, 10, 0, 40, 0, time.UTC), time.Date(2022, 2, 1, 10, 5, 40, 0, time.UTC))
	assert.NoError(t, err)

	timeframe, err := NewTimeframeParser("2022-02-01T10:00:40Z", "2022-02-01T10:05:40Z").Parse()
	assert.NoError(t, err)

	assert.EqualValues(t, expectedTimeframe.Start(), timeframe.Start())
	assert.EqualValues(t, expectedTimeframe.End(), timeframe.End())
}

func TestTimeframeParser_InvalidStart(t *testing.T) {
	timeframe, err := NewTimeframeParser("", "2022-02-01T10:05:40Z").Parse()
	assert.Error(t, err)
	assert.Nil(t, timeframe)
	assert.Contains(t, err.Error(), "error parsing timeframe start")
}

func TestTimeframeParser_InvalidEnd(t *testing.T) {
	timeframe, err := NewTimeframeParser("2022-02-01T10:00:40Z", "").Parse()
	assert.Error(t, err)
	assert.Nil(t, timeframe)
	assert.Contains(t, err.Error(), "error parsing timeframe end")
}
