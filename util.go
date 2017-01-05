package simulation

import (
	"hash/fnv"
	"math/rand"
	"time"
)

// GenerateWorkTime uses exprand to generate a reasonable work length
// on the order of milliseconds.
func GenerateWorkTime(averageWorkTime time.Duration) time.Duration {
	return time.Duration(rand.ExpFloat64() * float64(averageWorkTime))
}

// HashIdentifier hashes a string into a uniformly distributed uint32.
func HashIdentifier(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// ExplodeDuration returns all the constitent parts of a time.Duration.
func ExplodeDuration(duration time.Duration) (
	hours time.Duration,
	minutes time.Duration,
	seconds time.Duration,
	milliseconds time.Duration,
	microseconds time.Duration,
) {
	hours = duration / time.Hour
	hoursRemainder := duration - (hours * time.Hour)
	minutes = hoursRemainder / time.Minute
	minuteRemainder := hoursRemainder - (minutes * time.Minute)
	seconds = minuteRemainder / time.Second
	secondsRemainder := minuteRemainder - (seconds * time.Second)
	milliseconds = secondsRemainder / time.Millisecond
	millisecondsRemainder := secondsRemainder - (milliseconds * time.Millisecond)
	microseconds = millisecondsRemainder / time.Microsecond
	return
}

// RoundDuration rounds a duration to the given place.
func RoundDuration(duration, roundTo time.Duration) time.Duration {
	hours, minutes, seconds, milliseconds, microseconds := ExplodeDuration(duration)
	hours = hours * time.Hour
	minutes = minutes * time.Minute
	seconds = seconds * time.Second
	milliseconds = milliseconds * time.Millisecond
	microseconds = microseconds * time.Microsecond

	var total time.Duration
	if hours >= roundTo {
		total = total + hours
	}
	if minutes >= roundTo {
		total = total + minutes
	}
	if seconds >= roundTo {
		total = total + seconds
	}
	if milliseconds >= roundTo {
		total = total + milliseconds
	}
	if microseconds >= roundTo {
		total = total + microseconds
	}

	return total
}
