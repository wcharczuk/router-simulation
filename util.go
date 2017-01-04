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
