package internal

import "time"

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
		now = Time(time.Now())
	}
	return now.Round(time.Millisecond).UTC()
}
