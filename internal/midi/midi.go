package midi

import (
	"bytes"
	"os"
	"os/exec"
	"slices"

	gomidi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/gm"
	"gitlab.com/gomidi/midi/v2/smf"
)

type Midi struct {
	smf smf.SMF
}

func New(path string) (Midi, error) {
	parsedSMF, err := smf.ReadFile(path)
	if err != nil {
		return Midi{}, err
	}

	return Midi{smf: *parsedSMF}, nil
}

func (m Midi) getTrackIndexSet(trackIndexes ...int) (map[int]struct{}, bool) {
	trackIndexMap := make(map[int]struct{})
	for _, index := range trackIndexes {
		trackIndexMap[index] = struct{}{}
	}

	shouldIncludeAllTracks := len(trackIndexMap) == 0

	return trackIndexMap, shouldIncludeAllTracks
}

func (m *Midi) TrimByDuration(maxSeconds float64, trackIndexes ...int) {
	// Determine which tracks to process
	trackIndexMap, shouldIncludeAllTracks := m.getTrackIndexSet(trackIndexes...)

	threshold := int64(maxSeconds * 1_000_000) // convert seconds to microseconds

	for trIdx, tr := range m.smf.Tracks {
		// If not processing all tracks, skip this one if not in the map
		if !shouldIncludeAllTracks {
			if _, exists := trackIndexMap[trIdx]; !exists {
				continue
			}
		}

		var (
			newTrack []smf.Event
			absTicks int64
		)

		for _, ev := range tr {
			absTicks += int64(ev.Delta)
			// Get the event time in microseconds
			eventTimeMicro := m.smf.TimeAt(absTicks)

			// If the event's time is within the desired threshold, keep it
			if eventTimeMicro <= threshold {
				newTrack = append(newTrack, ev)
			} else {
				// Once we exceed the threshold, stop processing further events
				break
			}
		}

		// Ensure there's an End-Of-Track event if needed
		if len(newTrack) > 0 {
			lastEvent := newTrack[len(newTrack)-1]
			if lastEvent.Message.Type() != smf.MetaEndOfTrackMsg {
				// Add an End-Of-Track event with 0 delta
				newTrack = append(newTrack, smf.Event{
					Delta:   0,
					Message: smf.EOT,
				})
			}
		}

		// Update the track in the SMF
		m.smf.Tracks[trIdx] = newTrack
	}
}

func (m *Midi) RemoveLeadingSilence(trackIndexes ...int) {
	// Determine which tracks to process
	trackIndexMap, shouldIncludeAllTracks := m.getTrackIndexSet(trackIndexes...)

	// Find the minimal absolute tick of the first note-on event among the specified tracks
	var minTick int64 = -1

	for trIdx, tr := range m.smf.Tracks {
		// Skip tracks not in our set (if we're *not* including all tracks)
		if !shouldIncludeAllTracks {
			if _, exists := trackIndexMap[trIdx]; !exists {
				continue
			}
		}

		var absTick int64
		for _, ev := range tr {
			var vel uint8

			absTick += int64(ev.Delta)
			if ev.Message.GetNoteOn(nil, nil, &vel) && vel > 0 {
				if minTick == -1 || absTick < minTick {
					minTick = absTick
				}
				// Stop scanning this track after the first note-on
				break
			}
		}
	}

	// If minTick <= 0, there's either no note-on event or no leading silence to remove
	if minTick <= 0 {
		return
	}

	// Adjust each specified track's events to remove the initial silence
	for trIdx, tr := range m.smf.Tracks {
		// Skip tracks not in our set (if we're *not* including all tracks)
		if !shouldIncludeAllTracks {
			if _, exists := trackIndexMap[trIdx]; !exists {
				continue
			}
		}

		var newEvents []smf.Event
		var absTick int64
		var lastNewAbsTick int64

		for _, ev := range tr {
			absTick += int64(ev.Delta)

			// Shift this event's absolute tick by minTick
			newAbsTick := max(absTick-minTick, 0)

			// Recompute delta from the previous event's new absolute tick
			delta := max(newAbsTick-lastNewAbsTick, 0)

			newEvents = append(newEvents, smf.Event{
				Delta:   uint32(delta),
				Message: ev.Message,
			})

			lastNewAbsTick = newAbsTick
		}

		// Update the track with the adjusted events
		m.smf.Tracks[trIdx] = newEvents
	}
}

func (m *Midi) RemoveTracks(trackIndexes ...int) {
	trackIndexMap, shouldIncludeAllTracks := m.getTrackIndexSet(trackIndexes...)

	var remainingTracks []smf.Track

	for trackIndex, track := range m.smf.Tracks {
		if !shouldIncludeAllTracks {
			if _, exists := trackIndexMap[trackIndex]; !exists {
				remainingTracks = append(remainingTracks, track)
			}
		}
	}

	m.smf.Tracks = remainingTracks
}

func (m Midi) ExtractNoteOnTimestamps(trackIndexes ...int) []float64 {
	trackIndexMap, shouldIncludeAllTracks := m.getTrackIndexSet(trackIndexes...)

	var noteOnTimestamps []float64

	for trIdx, tr := range m.smf.Tracks {
		if !shouldIncludeAllTracks {
			if _, exists := trackIndexMap[trIdx]; !exists {
				continue
			}
		}

		var absTicks int64

		for _, ev := range tr {
			absTicks += int64(ev.Delta)

			// Skip if multiple events have the same absolute tick
			if ev.Delta == 0 {
				continue
			}

			if ev.Message.GetNoteStart(nil, nil, nil) {
				noteStartMicro := m.smf.TimeAt(absTicks) // returns int64 microseconds
				noteStartSeconds := float64(noteStartMicro) / 1_000_000.0
				noteOnTimestamps = append(noteOnTimestamps, noteStartSeconds)
			}
		}
	}

	// If multiple tracks are processed, sort the timestamps
	if len(m.smf.Tracks) > 1 && (shouldIncludeAllTracks || len(trackIndexes) > 1) {
		slices.Sort(noteOnTimestamps)
	}

	return noteOnTimestamps
}

func (m *Midi) SetInstrument(instrument gm.Instr, trackIndexes ...int) {
	trackIndexMap, shouldIncludeAllTracks := m.getTrackIndexSet(trackIndexes...)

	for trIdx, tr := range m.smf.Tracks {
		if !shouldIncludeAllTracks {
			if _, exists := trackIndexMap[trIdx]; !exists {
				continue
			}
		}

		var usedChannels []uint8
		var programChangeEventIndices []int

		// Find all program change events and collect the channel numbers
		for eventIndex, event := range tr {
			var channel uint8

			if event.Message.GetProgramChange(&channel, nil) {
				usedChannels = append(usedChannels, channel)
				programChangeEventIndices = append(programChangeEventIndices, eventIndex)
			}
		}

		// Remove all program change events, accounting for index shifting
		for i, eventIndex := range programChangeEventIndices {
			adjustedIndex := eventIndex - i
			tr = slices.Delete(tr, adjustedIndex, adjustedIndex+1)
		}

		// Add new program change events at the beginning of the track
		for _, channel := range usedChannels {
			programChangeEvent := smf.Event{
				Delta:   0,
				Message: smf.Message(gomidi.ProgramChange(channel, instrument.Value())),
			}
			tr = append([]smf.Event{programChangeEvent}, tr...)
		}

		m.smf.Tracks[trIdx] = tr
	}
}

func (m Midi) SaveMidi(filePath string) error {
	return m.smf.WriteFile(filePath)
}

func (m Midi) ToBytes() ([]byte, error) {
	var buffer bytes.Buffer

	_, err := m.smf.WriteTo(&buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (m Midi) ToWav(soundFontPath string) (string, error) {
	midFilePath := os.TempDir() + `\_.mid`
	wavFilePath := os.TempDir() + `\_.wav`

	m.SaveMidi(midFilePath)

	cmd := exec.Command(
		"fluidsynth",
		"-ni",
		soundFontPath,
		midFilePath,
		"-F", wavFilePath,
		"-r 44100",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return wavFilePath, nil
}

func FilterTimestampsByCloseness(timestamps []float64, closenessThresholdMs int32) []float64 {
	if closenessThresholdMs <= 0 || len(timestamps) == 0 {
		return timestamps
	}

	closenessThresholdSec := float64(closenessThresholdMs) / 1000
	filteredTimestamps := []float64{timestamps[0]}
	lastAcceptedTime := timestamps[0]

	for _, currentTimestamp := range timestamps[1:] {
		if currentTimestamp-lastAcceptedTime >= closenessThresholdSec {
			filteredTimestamps = append(filteredTimestamps, currentTimestamp)
			lastAcceptedTime = currentTimestamp
		}
	}

	return filteredTimestamps
}
