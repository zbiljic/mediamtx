package recorder

import (
	"time"
)

// a round time is skipped when it is closer than this to the start of a segment.
// This prevents the creation of a tiny segment when a segment starts slightly before a round time,
// which happens when the first sample of a segment belongs to a track that is behind the one
// that caused the segment switch, or when the segment switch is caused by the segment duration.
// It corresponds to the maximum distance between the tracks of a segment (maxBasetime).
const roundTolerance = 1 * time.Second

// nextRoundTime returns the first time after t that is a multiple of d in the wall clock of loc.
func nextRoundTime(t time.Time, d time.Duration, loc *time.Location) time.Time {
	_, offsetSecs := t.In(loc).Zone()
	offset := time.Duration(offsetSecs) * time.Second

	// time.Truncate() works on absolute time, shift the time in order to work on the wall clock.
	next := t.Add(offset).Truncate(d).Add(d).Add(-offset)

	// the UTC offset might change between t and the round time (DST).
	// Realign the round time if it is not a multiple of d in the new wall clock.
	_, nextOffsetSecs := next.In(loc).Zone()
	if diff := time.Duration(nextOffsetSecs)*time.Second - offset; diff%d != 0 {
		if realigned := next.Add(-diff); realigned.After(t) {
			next = realigned
		}
	}

	return next
}

// segmentRoundEnd returns the time after which a segment has to be closed
// in order to end at a multiple of d. ref is the time of the sample that opened the segment.
// It returns the zero time when rounding is disabled.
func segmentRoundEnd(ref time.Time, d time.Duration, loc *time.Location) time.Time {
	if d <= 0 || ref.IsZero() {
		return time.Time{}
	}

	return nextRoundTime(ref.Add(min(roundTolerance, d/2)), d, loc)
}
