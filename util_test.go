package main

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
)

func TestGenerateArrivalTimeInterval(t *testing.T) {
	assert := assert.New(t)

	assert.InDelta(float64(generateArrivalTimeInterval()), 0.0, 30.0*float64(time.Millisecond))
}
