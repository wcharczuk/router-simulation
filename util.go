package main

import (
	"math/rand"
	"time"

	util "github.com/blendlabs/go-util"
)

// generateArrivalTimeDelta generates the delta from now to when
// a client arrives into the simulation on the order of milliseconds.
func generateArrivalTimeInterval() time.Duration {
	randMillis := rand.ExpFloat64() * 2
	return time.Duration(randMillis) * time.Millisecond
}

// generateWorkTime uses exprand to generate a reasonable work length
// on the order of milliseconds.
func generateWorkTime() time.Duration {
	randMillis := rand.ExpFloat64() * 30
	return time.Duration(randMillis) * time.Millisecond
}

// generateID returns a new random, 8 character, identifier
func generateID() string {
	return util.String.RandomString(8)
}
