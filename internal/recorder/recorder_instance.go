package recorder

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/recordstore"
	"github.com/bluenviron/mediamtx/internal/stream"
)

type recorderInstance struct {
	pathFormat           string
	format               conf.RecordFormat
	partDuration         time.Duration
	maxPartSize          conf.StringSize
	segmentDuration      time.Duration
	segmentRoundDuration time.Duration
	startRoundDuration   time.Duration
	pathName             string
	stream               *stream.Stream
	onSegmentCreate      OnSegmentCreateFunc
	onSegmentComplete    OnSegmentCompleteFunc
	location             *time.Location
	parent               logger.Writer

	streamID    uuid.UUID
	pathFormat2 string
	format2     format
	skip        bool
	reader      *stream.Reader

	// in reader goroutine
	startAfterNTP time.Time

	terminate chan struct{}
	done      chan struct{}
}

// Log implements logger.Writer.
func (ri *recorderInstance) Log(level logger.Level, format string, args ...any) {
	ri.parent.Log(level, format, args...)
}

// discardBeforeStart returns whether a sample has to be discarded in order to start
// the recording at a multiple of startRoundDuration.
// It must be called until the first segment is created.
func (ri *recorderInstance) discardBeforeStart(ntp time.Time, hasVideo bool, isVideo bool, randomAccess bool) bool {
	if ri.startRoundDuration <= 0 || ntp.IsZero() {
		return false
	}

	if ri.startAfterNTP.IsZero() {
		ri.startAfterNTP = nextRoundTime(ntp, ri.startRoundDuration, ri.location)
	}

	if ntp.Before(ri.startAfterNTP) {
		return true
	}

	// samples before the round time have been discarded,
	// therefore the recording must start from a random access sample of the video track, if any.
	return hasVideo && (!isVideo || !randomAccess)
}

func (ri *recorderInstance) initialize() {
	ri.streamID = uuid.New()
	ri.pathFormat2 = ri.pathFormat
	ri.pathFormat2 = recordstore.PathAddExtension(
		strings.ReplaceAll(ri.pathFormat2, "%path", ri.pathName),
		ri.format,
	)
	ri.reader = &stream.Reader{
		SkipOutboundBytes: true,
		Parent:            ri,
	}

	ri.terminate = make(chan struct{})
	ri.done = make(chan struct{})

	switch ri.format {
	case conf.RecordFormatMPEGTS:
		ri.format2 = &formatMPEGTS{
			ri: ri,
		}
		ok := ri.format2.initialize()
		ri.skip = !ok

	default:
		ri.format2 = &formatFMP4{
			ri: ri,
		}
		ok := ri.format2.initialize()
		ri.skip = !ok
	}

	if !ri.skip {
		ri.stream.AddReader(ri.reader)
	}

	go ri.run()
}

func (ri *recorderInstance) close() {
	close(ri.terminate)
	<-ri.done
}

func (ri *recorderInstance) run() {
	defer close(ri.done)

	if !ri.skip {
		select {
		case err := <-ri.reader.Error():
			ri.Log(logger.Error, err.Error())

		case <-ri.terminate:
		}

		ri.stream.RemoveReader(ri.reader)
	} else {
		<-ri.terminate
	}

	ri.format2.close()
}
