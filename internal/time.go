package internal

import (
	"fmt"
	"time"
)

// CurrentTimestamp is *the* way to get a current timestamps in OTF and
// time.Now() should be avoided.
//
// We want timestamps to be rounded to nearest
// millisecond so that they can be persisted/serialised and not lose precision
// thereby making comparisons and testing easier.
//
// We also want timestamps to be in the UTC time zone. Again it makes
// testing easier because libs such as testify's assert use DeepEqual rather
// than time.Equal to compare times (and structs containing times). That means
// the internal representation is compared, including the time zone which may
// differ even though two times refer to the same instant.
//
// In any case, the time zone of the server is often not of importance, whereas
// that of the user often is, and conversion to their time zone is necessary
// regardless.
//
// And the optional now arg gives tests the opportunity to swap out time.Now() with
// a deterministic time. If it's nil then time.Now() is used.
func CurrentTimestamp(now *time.Time) time.Time {
	if now == nil {
		now = new(time.Now())
	}
	return now.Round(time.Millisecond).UTC()
}

func Ago(now, t time.Time) string {
	diff := now.Sub(t)
	var (
		n      int
		suffix string
	)

	switch {
	// If less than 10 seconds, then report number of seconds ago
	case diff < time.Second*10:
		n = int(diff.Seconds())
		suffix = "s"
		// If between 10 seconds and a minute then report number of seconds in 10
		// second blocks. We do this because this func is called on every render,
		// and it can be discombobulating to the user when every row in a table is
		// updating every second as they navigate it...using 10 seconds blocks helps
		// a little.
	case diff < time.Minute:
		n = int(diff.Round(10 * time.Second).Seconds())
		suffix = "s"
	case diff < time.Hour:
		n = int(diff.Minutes())
		suffix = "m"
	case diff < 24*time.Hour:
		n = int(diff.Hours())
		suffix = "h"
	default:
		n = int(diff.Hours() / 24)
		suffix = "d"
	}
	return fmt.Sprintf("%d%s ago", n, suffix)
}
