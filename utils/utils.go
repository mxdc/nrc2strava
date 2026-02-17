package utils

import (
	"time"
)

func ParseTimeInMs(epochMs int64) time.Time {
	// Convert milliseconds to seconds and nanoseconds
	seconds := epochMs / 1000
	nanoseconds := (epochMs % 1000) * 1_000_000

	// Parse the timestamp
	parsedTime := time.Unix(seconds, nanoseconds).UTC()

	// Print the parsed time
	return parsedTime
}
