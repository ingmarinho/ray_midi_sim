package sim

import (
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Interval struct {
	start float32
	end   float32
}

// addInterval unions `newIv` into the slice of Intervals, returning an updated slice.
// We store them sorted by start and then merge any overlapping Intervals.
func addInterval(Intervals []Interval, newIv Interval) []Interval {
	// Ignore empty Intervals
	if newIv.end <= newIv.start {
		return Intervals
	}

	// 1. Append, then sort
	Intervals = append(Intervals, newIv)
	// Simple sort by .start
	sort.Slice(Intervals, func(i, j int) bool {
		return Intervals[i].start < Intervals[j].start
	})

	// 2. Merge pass
	merged := make([]Interval, 0, len(Intervals))
	current := Intervals[0]
	for i := 1; i < len(Intervals); i++ {
		next := Intervals[i]
		// Overlap?
		if next.start <= current.end {
			// Extend the current Interval's end if needed
			if next.end > current.end {
				current.end = next.end
			}
		} else {
			// No overlap, push current and move on
			merged = append(merged, current)
			current = next
		}
	}
	// Push the final
	merged = append(merged, current)

	return merged
}

// subtractInterval removes the segment [skipStart, skipEnd] from the given Intervals.
// The result is a new set of Intervals that do NOT include [skipStart, skipEnd].
func subtractInterval(Intervals []Interval, skipStart, skipEnd float32) []Interval {
	var output []Interval

	for _, iv := range Intervals {
		// If there's no overlap, keep the Interval as-is.
		if skipEnd <= iv.start || skipStart >= iv.end {
			// No intersection
			output = append(output, iv)
			continue
		}

		// If there is an overlap, break it into up to two pieces.
		// Left side if skipStart > iv.start
		if skipStart > iv.start {
			output = append(output, Interval{iv.start, skipStart})
		}
		// Right side if skipEnd < iv.end
		if skipEnd < iv.end {
			output = append(output, Interval{skipEnd, iv.end})
		}
	}

	return output
}

//------------------------------------------------------------------------------
// Helper functions to gather vertical/horizontal intervals from a list of rects
//------------------------------------------------------------------------------

// gatherVerticalIntervals collects all intervals [rect.Y, rect.Y+rect.Height]
// where x is inside a rect's [X, X+Width], clipped to [startY, endY], and merges them.
func gatherVerticalIntervals(x, startY, endY float32, rects []rl.Rectangle) []Interval {
	intervals := []Interval{}
	for _, r := range rects {
		// Check if the vertical line x is within the rect horizontally
		if x > r.X && x < r.X+r.Width {
			top := r.Y
			bot := r.Y + r.Height

			// Clip to [startY, endY]
			if bot < startY || top > endY {
				continue
			}
			if top < startY {
				top = startY
			}
			if bot > endY {
				bot = endY
			}

			// Merge this interval into our list
			intervals = addInterval(intervals, Interval{start: top, end: bot})
		}
	}
	return intervals
}

// gatherHorizontalIntervals collects all intervals [rect.X, rect.X+rect.Width]
// where y is inside a rect's [Y, Y+Height], clipped to [startX, endX], and merges them.
func gatherHorizontalIntervals(y, startX, endX float32, rects []rl.Rectangle) []Interval {
	intervals := []Interval{}
	for _, r := range rects {
		// Check if the horizontal line y is within the rect vertically
		if y > r.Y && y < r.Y+r.Height {
			left := r.X
			right := r.X + r.Width

			// Clip to [startX, endX]
			if right < startX || left > endX {
				continue
			}
			if left < startX {
				left = startX
			}
			if right > endX {
				right = endX
			}

			// Merge this interval into our list
			intervals = addInterval(intervals, Interval{start: left, end: right})
		}
	}
	return intervals
}
