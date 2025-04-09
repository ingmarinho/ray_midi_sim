package sim

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Cell struct {
	pos  rl.Vector2
	size int
}

func NewCell(x, y float32, size int) Cell {
	return Cell{
		pos:  rl.NewVector2(x, y),
		size: size,
	}
}

func (c Cell) ToRect() rl.Rectangle {
	return rl.NewRectangle(c.pos.X, c.pos.Y, float32(c.size), float32(c.size))
}

func cellToPixelRect(cx, cy, w, h int) rl.Rectangle {
	return rl.Rectangle{
		X:      float32(cx * CELL_SIZE),
		Y:      float32(cy * CELL_SIZE),
		Width:  float32(w * CELL_SIZE),
		Height: float32(h * CELL_SIZE),
	}
}

// findReachableCells finds all grid‐cells where a rectangle of dimensions rect could be placed
// without overlapping safeAreas and within maxCellDistance of the start.
//
// rect is the *starting rectangle* in PIXEL space (for example, 3x2 cells at some position).
// safeAreas are also in PIXEL space (e.g., walls).
// maxCellDistance is in *cells*. (If it’s in pixels, you’d handle it differently.)
//
// This function returns a slice of rl.Vector2, where each Vector2 is the top‐left corner
// of a valid rectangle position *in pixel coordinates*.
func findReachableCells(rect rl.Rectangle, safeAreas []rl.Rectangle, maxCellDistance int) []Cell {
	// 1) (Optional) Build a "max expansion" rectangle in pixel coordinates
	//    so we can skip checking collisions against walls that are obviously out of range.
	//    If maxCellDistance is in *cells*, multiply by CELL_SIZE:
	maxDistancePx := float32(maxCellDistance * CELL_SIZE)
	maxExpansionRect := rl.NewRectangle(
		rect.X-maxDistancePx,
		rect.Y-maxDistancePx,
		rect.Width+maxDistancePx*2,
		rect.Height+maxDistancePx*2,
	)

	// Collect only relevant safeAreas (within the bounding box).
	relevantSafeAreas := make([]rl.Rectangle, 0, len(safeAreas))
	for _, sa := range safeAreas {
		if rl.CheckCollisionRecs(maxExpansionRect, sa) {
			relevantSafeAreas = append(relevantSafeAreas, sa)
		}
	}

	// 2) Convert the input "start rect" to *cell* coordinates.
	//    The BFS will treat (startCx, startCy) as the top-left cell of the rectangle.
	startCx := int(rect.X) / CELL_SIZE
	startCy := int(rect.Y) / CELL_SIZE

	// Also figure out how many cells wide/tall this rect is.
	widthInCells := int(rect.Width) / CELL_SIZE
	heightInCells := int(rect.Height) / CELL_SIZE

	// 3) BFS data structures
	type queueEntry struct {
		cx, cy int // which cell we occupy (top-left corner of the rectangle in cell coords)
		dist   int // distance in "cells" from the start
	}
	visited := make(map[[2]int]bool)
	var result []Cell

	// Begin BFS from the start cell
	queue := []queueEntry{{cx: startCx, cy: startCy, dist: 0}}
	visited[[2]int{startCx, startCy}] = true

	// 4) BFS neighbor enqueuing
	tryEnqueue := func(cx, cy, dist int) {
		// If we've gone beyond the maximum allowed distance (in cells), skip.
		if dist > maxCellDistance {
			return
		}
		key := [2]int{cx, cy}
		if !visited[key] {
			// Convert this cell position to a rectangle in pixel space
			checkRect := cellToPixelRect(cx, cy, widthInCells, heightInCells)
			// If it doesn't collide with any relevant walls, add to queue
			if !collidesWithAny(checkRect, relevantSafeAreas) {
				visited[key] = true
				queue = append(queue, queueEntry{cx: cx, cy: cy, dist: dist})
			}
		}
	}

	// 5) BFS loop
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		cx, cy, dist := current.cx, current.cy, current.dist

		// **Here** we store the top‐left corner in *pixel* coordinates.
		px := float32(cx * CELL_SIZE)
		py := float32(cy * CELL_SIZE)
		result = append(result, NewCell(px, py, CELL_SIZE))

		// Enqueue 4 neighbors: up, down, left, right (1 cell away)
		tryEnqueue(cx, cy-1, dist+1)
		tryEnqueue(cx, cy+1, dist+1)
		tryEnqueue(cx-1, cy, dist+1)
		tryEnqueue(cx+1, cy, dist+1)
	}

	return result
}
