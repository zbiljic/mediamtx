package recorder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNextRoundTime(t *testing.T) {
	kolkata, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)

	rome, err := time.LoadLocation("Europe/Rome")
	require.NoError(t, err)

	for _, ca := range []struct {
		name string
		t    time.Time
		d    time.Duration
		loc  *time.Location
		next time.Time
	}{
		{
			"utc",
			time.Date(2008, 5, 20, 22, 15, 25, 300000000, time.UTC),
			10 * time.Second,
			time.UTC,
			time.Date(2008, 5, 20, 22, 15, 30, 0, time.UTC),
		},
		{
			"already rounded",
			time.Date(2008, 5, 20, 22, 0, 0, 0, time.UTC),
			1 * time.Hour,
			time.UTC,
			time.Date(2008, 5, 20, 23, 0, 0, 0, time.UTC),
		},
		{
			"half hour offset",
			time.Date(2008, 5, 20, 22, 15, 25, 0, kolkata),
			1 * time.Hour,
			kolkata,
			time.Date(2008, 5, 20, 23, 0, 0, 0, kolkata),
		},
		{
			"day",
			time.Date(2008, 5, 20, 22, 15, 25, 0, rome),
			24 * time.Hour,
			rome,
			time.Date(2008, 5, 21, 0, 0, 0, 0, rome),
		},
		{
			"day, different location",
			time.Date(2008, 5, 20, 22, 15, 25, 0, time.UTC), // 00:15:25 in Rome
			24 * time.Hour,
			rome,
			time.Date(2008, 5, 22, 0, 0, 0, 0, rome),
		},
		{
			"day, DST start",
			time.Date(2008, 3, 30, 0, 30, 0, 0, rome),
			24 * time.Hour,
			rome,
			time.Date(2008, 3, 31, 0, 0, 0, 0, rome),
		},
		{
			"day, DST end",
			time.Date(2008, 10, 26, 0, 30, 0, 0, rome),
			24 * time.Hour,
			rome,
			time.Date(2008, 10, 27, 0, 0, 0, 0, rome),
		},
		{
			"hour, DST end",
			time.Date(2008, 10, 26, 0, 30, 0, 0, time.UTC), // 02:30 CEST
			1 * time.Hour,
			rome,
			time.Date(2008, 10, 26, 1, 0, 0, 0, time.UTC), // 02:00 CET
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			next := nextRoundTime(ca.t, ca.d, ca.loc)
			require.True(t, ca.next.Equal(next), "expected %v, got %v", ca.next, next.In(ca.loc))
		})
	}
}

func TestSegmentRoundEnd(t *testing.T) {
	ref := time.Date(2008, 5, 20, 22, 15, 25, 0, time.UTC)

	// disabled
	require.True(t, segmentRoundEnd(ref, 0, time.UTC).IsZero())
	require.True(t, segmentRoundEnd(time.Time{}, 10*time.Second, time.UTC).IsZero())

	require.Equal(t,
		time.Date(2008, 5, 20, 22, 15, 30, 0, time.UTC),
		segmentRoundEnd(ref, 10*time.Second, time.UTC))

	// round times that are too close to the start are skipped
	require.Equal(t,
		time.Date(2008, 5, 20, 22, 15, 40, 0, time.UTC),
		segmentRoundEnd(ref.Add(4990*time.Millisecond), 10*time.Second, time.UTC))

	// tolerance is proportional to short durations
	require.Equal(t,
		time.Date(2008, 5, 20, 22, 15, 26, 0, time.UTC),
		segmentRoundEnd(ref.Add(400*time.Millisecond), 1*time.Second, time.UTC))
}
