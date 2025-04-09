package sim

import (
	"math"

	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// IsAdjacent checks if two rectangles are adjacent (touching or nearly touching)
func IsAdjacent(r1, r2 rl.Rectangle, threshold float32) bool {
	// Check horizontal adjacency
	horizontal := (r1.X+r1.Width >= r2.X-threshold && r1.X <= r2.X+r2.Width+threshold) &&
		(r1.Y < r2.Y+r2.Height && r1.Y+r1.Height > r2.Y)

	// Check vertical adjacency
	vertical := (r1.Y+r1.Height >= r2.Y-threshold && r1.Y <= r2.Y+r2.Height+threshold) &&
		(r1.X < r2.X+r2.Width && r1.X+r1.Width > r2.X)

	return horizontal || vertical
}

// AdjustRectanglesOverlapping modifies the list of rectangles so that adjacent ones slightly overlap
func AdjustRectanglesOverlapping(rects []rl.Rectangle, margin float32) []rl.Rectangle {
	adjustedRects := make([]rl.Rectangle, len(rects))
	copy(adjustedRects, rects)

	for i := 0; i < len(adjustedRects); i++ {
		for j := 0; j < len(adjustedRects); j++ {
			if i == j {
				continue
			}

			if IsAdjacent(adjustedRects[i], adjustedRects[j], margin) {
				// Adjust horizontally if they are next to each other
				if adjustedRects[i].X < adjustedRects[j].X {
					adjustedRects[j].X = adjustedRects[i].X + adjustedRects[i].Width - margin
				} else if adjustedRects[i].X > adjustedRects[j].X {
					adjustedRects[i].X = adjustedRects[j].X + adjustedRects[j].Width - margin
				}

				// Adjust vertically if they are stacked
				if adjustedRects[i].Y < adjustedRects[j].Y {
					adjustedRects[j].Y = adjustedRects[i].Y + adjustedRects[i].Height - margin
				} else if adjustedRects[i].Y > adjustedRects[j].Y {
					adjustedRects[i].Y = adjustedRects[j].Y + adjustedRects[j].Height - margin
				}
			}
		}
	}

	return adjustedRects
}

// collidesWithAny returns true if rect intersects with any rectangle in rects.
func collidesWithAny(rect rl.Rectangle, rects []rl.Rectangle) bool {
	for _, r := range rects {
		if rl.CheckCollisionRecs(rect, r) {
			return true
		}
	}
	return false
}

func mergeNeighboringRects(rects []rl.Rectangle) []rl.Rectangle {
	// First, filter out any rectangles that have zero or negative width/height.
	var filtered []rl.Rectangle
	for _, r := range rects {
		if r.Width <= 0 || r.Height <= 0 {
			continue
		}
		filtered = append(filtered, r)
	}

	// We'll iteratively merge rectangles until no more merges can be done.
	mergedRects := slices.Clone(filtered)

	for {
		mergedAnything := false

		// Attempt to merge pairs of rectangles
	outer:
		for i := range mergedRects {
			for j := i + 1; j < len(mergedRects); j++ {
				if isNeighboring(mergedRects[i], mergedRects[j], 2) {
					// Merge them into bounding rect
					newRect := mergeRect(mergedRects[i], mergedRects[j])

					// Remove the two old ones from the slice
					mergedRects = slices.Delete(mergedRects, i, i+1)
					// After removal at i, j shifts left by 1
					j--
					mergedRects = slices.Delete(mergedRects, j, j+1)

					// Add the merged rectangle
					mergedRects = append(mergedRects, newRect)

					mergedAnything = true
					// Break out so we start from scratch again
					break outer
				}
			}
		}

		if !mergedAnything {
			break
		}
	}

	return mergedRects
}

func isNeighboring(a, b rl.Rectangle, offset float32) bool {
	// Check horizontal adjacency
	horizontal := (a.X+a.Width >= b.X-offset && a.X <= b.X+b.Width+offset) &&
		(a.Y < b.Y+b.Height && a.Y+a.Height > b.Y)

	// Check vertical adjacency
	vertical := (a.Y+a.Height >= b.Y-offset && a.Y <= b.Y+b.Height+offset) &&
		(a.X < b.X+b.Width && a.X+a.Width > b.X)

	return horizontal || vertical
}

// MergeOverlappingRects takes a list of rectangles and merges any that overlap
// or share an edge. It returns a new slice of merged rectangles.
func mergeOverlappingRects(rects []rl.Rectangle) []rl.Rectangle {
	// First, filter out any rectangles that have zero or negative width/height.
	var filtered []rl.Rectangle
	for _, r := range rects {
		if r.Width <= 0 || r.Height <= 0 {
			continue
		}
		filtered = append(filtered, r)
	}

	// We'll iteratively merge rectangles until no more merges can be done.
	mergedRects := slices.Clone(filtered)

	for {
		mergedAnything := false

		// Attempt to merge pairs of rectangles
	outer:
		for i := range mergedRects {
			for j := i + 1; j < len(mergedRects); j++ {
				if canMerge(mergedRects[i], mergedRects[j]) {
					// Merge them into bounding rect
					newRect := mergeRect(mergedRects[i], mergedRects[j])

					// Remove the two old ones from the slice
					mergedRects = slices.Delete(mergedRects, i, i+1)
					// After removal at i, j shifts left by 1
					j--
					mergedRects = slices.Delete(mergedRects, j, j+1)

					// Add the merged rectangle
					mergedRects = append(mergedRects, newRect)

					mergedAnything = true
					// Break out so we start from scratch again
					break outer
				}
			}
		}

		if !mergedAnything {
			break
		}
	}

	return mergedRects
}

// canMerge returns true if two rectangles overlap or share an edge (meaning
// their bounding rectangle has an area equal to or less than the sum of their areas).
func canMerge(a, b rl.Rectangle) bool {
	// Compute bounding rectangle
	merged := mergeRect(a, b)

	areaA := a.Width * a.Height
	areaB := b.Width * b.Height
	areaMerged := merged.Width * merged.Height

	// If the bounding rectangle’s area equals (or is slightly larger/equals) the sum
	// of the areas, they at least touch or overlap fully.
	// Use <= if you want to allow edges to be considered “touching” merges.
	return areaMerged <= areaA+areaB
}

func mergeRect(rectA, rectB rl.Rectangle) rl.Rectangle {
	minX := float32(math.Min(float64(rectA.X), float64(rectB.X)))
	minY := float32(math.Min(float64(rectA.Y), float64(rectB.Y)))
	maxX := float32(math.Max(float64(rectA.X+rectA.Width), float64(rectB.X+rectB.Width)))
	maxY := float32(math.Max(float64(rectA.Y+rectA.Height), float64(rectB.Y+rectB.Height)))

	return rl.NewRectangle(minX, minY, maxX-minX, maxY-minY)
}
