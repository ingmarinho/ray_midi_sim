package sim

import (
	"errors"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Map struct {
	bounces              []Bounce
	floatingBounceRects  []rl.Rectangle
	connectedBounceRects []rl.Rectangle
	safeAreas            []rl.Rectangle

	polygonPaths []Polygon
}

func (m Map) Bounces() []Bounce {
	return m.bounces
}

func (m *Map) PopBounce() Bounce {
	b := m.bounces[0]
	m.bounces = m.bounces[1:]
	return b
}

func (m *Map) Draw(cameraRect rl.Rectangle) {
	// for _, r := range m.safeAreas {
	// 	if !rl.CheckCollisionRecs(cameraRect, r) {
	// 		continue
	// 	}

	// 	rl.DrawRectangleRec(r, rl.White)
	// 	rl.DrawRectangleLinesEx(r, 1, rl.Black)
	// }
	// for _, b := range m.bounces {
	// 	if !rl.CheckCollisionRecs(cameraRect, b.ToRect()) {
	// 		continue
	// 	}

	// 	b.Draw()
	// }

	// draw path polygons using lines
	// for _, pp := range m.polygonPaths {
	// 	for i := 0; i < len(pp)-1; i++ {
	// 		rl.DrawLineV(pp[i], pp[i+1], rl.Black)
	// 	}

	// 	// draw last line
	// 	rl.DrawLineV(pp[len(pp)-1], pp[0], rl.Black)
	// }
}

func createPathPolygon(direction, startPos, endPos rl.Vector2, squareSize int) []rl.Vector2 {
	switch direction {
	case rl.Vector2{X: 1, Y: 1}:
		// moving towards bottom right
		return []rl.Vector2{
			// top left of (start) square
			rl.NewVector2(startPos.X, startPos.Y),
			// top right of (start) square
			rl.NewVector2(startPos.X+float32(squareSize), startPos.Y),
			// top right of (end) square
			rl.NewVector2(endPos.X+float32(squareSize), endPos.Y),
			// bottom right of (end) square
			rl.NewVector2(endPos.X+float32(squareSize), endPos.Y+float32(squareSize)),
			// bottom left of (end) square
			rl.NewVector2(endPos.X, endPos.Y+float32(squareSize)),
			// bottom left of (start) square
			rl.NewVector2(startPos.X, startPos.Y+float32(squareSize)),
		}
	case rl.Vector2{X: -1, Y: 1}:
		// moving towards bottom left
		return []rl.Vector2{
			// top right of (start) square
			rl.NewVector2(startPos.X+float32(squareSize), startPos.Y),
			// top left of (start) square
			rl.NewVector2(startPos.X, startPos.Y),
			// top left of (end) square
			rl.NewVector2(endPos.X, endPos.Y),
			// bottom left of (end) square
			rl.NewVector2(endPos.X, endPos.Y+float32(squareSize)),
			// bottom right of (end) square
			rl.NewVector2(endPos.X+float32(squareSize), endPos.Y+float32(squareSize)),
			// bottom right of (start) square
			rl.NewVector2(startPos.X+float32(squareSize), startPos.Y+float32(squareSize)),
		}

	case rl.Vector2{X: 1, Y: -1}:
		// moving towards top right
		return []rl.Vector2{
			// bottom left of (start) square
			rl.NewVector2(startPos.X, startPos.Y+float32(squareSize)),
			// bottom right of (start) square
			rl.NewVector2(startPos.X+float32(squareSize), startPos.Y+float32(squareSize)),
			// bottom right of (end) square
			rl.NewVector2(endPos.X+float32(squareSize), endPos.Y+float32(squareSize)),
			// top right of (end) square
			rl.NewVector2(endPos.X+float32(squareSize), endPos.Y),
			// top left of (end) square
			rl.NewVector2(endPos.X, endPos.Y),
			// top left of (start) square
			rl.NewVector2(startPos.X, startPos.Y),
		}
	case rl.Vector2{X: -1, Y: -1}:
		// moving towards top left
		return []rl.Vector2{
			// bottom right of (start) square
			rl.NewVector2(startPos.X+float32(squareSize), startPos.Y+float32(squareSize)),
			// bottom left of (start) square
			rl.NewVector2(startPos.X, startPos.Y+float32(squareSize)),
			// bottom left of (end) square
			rl.NewVector2(endPos.X, endPos.Y+float32(squareSize)),
			// top left of (end) square
			rl.NewVector2(endPos.X, endPos.Y),
			// top right of (end) square
			rl.NewVector2(endPos.X+float32(squareSize), endPos.Y),
			// top right of (start) square
			rl.NewVector2(startPos.X+float32(squareSize), startPos.Y),
		}
	}
	// unreachable
	return []rl.Vector2{}
}

func rectCornersCollideWithPolygon(polygon []rl.Vector2, rect rl.Rectangle) bool {
	corners := []rl.Vector2{
		rl.NewVector2(rect.X, rect.Y),
		rl.NewVector2(rect.X+rect.Width, rect.Y),
		rl.NewVector2(rect.X+rect.Width, rect.Y+rect.Height),
		rl.NewVector2(rect.X, rect.Y+rect.Height),
	}

	for _, corner := range corners {
		if rl.CheckCollisionPointPoly(corner, polygon) {
			return true
		}
	}

	return false
}

func GenerateMap(noteOnTimestamps []float64, squareSpeed int) (Map, error) {
	var recursiveGenerate func(square Square, noteOnTimestamps []float64, depth int, bounces []Bounce, prevTime float64, prevBounceDirPriority [2]BounceDirection) []Bounce

	m := Map{}

	var safeAreas []rl.Rectangle
	var polygonPaths []Polygon

	backtrackSteps := 0

	recursiveGenerate = func(square Square, noteOnTimestamps []float64, depth int, bounces []Bounce, prevTimeSec float64, prevBounceDirPriority [2]BounceDirection) []Bounce {
		if len(noteOnTimestamps) < 1 {

			return bounces
		}

		noteTimeSec := noteOnTimestamps[0]
		dt := noteTimeSec - prevTimeSec

		polygonPathsStartIndex := len(polygonPaths)

		prevSquareRect := square.ToRectangle()
		prevPos := square.position

		square.Update(float32(dt))

		// snap position
		snappedPos := snapPosition(square.position, float32(CELL_SIZE), float32(CELL_SIZE))
		square.position = snappedPos

		polygonPath := createPathPolygon(square.direction, prevPos, snappedPos, SQUARE_SIZE)
		polygonPaths = append(polygonPaths, polygonPath)

		// check if any bounces exist, if so, check for collisions
		if len(bounces) > 0 {
			collisionBounceRects := make([]rl.Rectangle, 0, len(bounces))
			for _, bounce := range bounces {
				collisionBounceRects = append(collisionBounceRects, bounce.ToCollisionRect())
			}

			// path collision check
			pathCollision := false
			for _, pp := range polygonPaths {
				if rectCornersCollideWithPolygon(pp, collisionBounceRects[len(collisionBounceRects)-1]) {
					pathCollision = true
					break
				}
			}

			// bounce rect collision check
			bounceRectCollision := false
			for _, bounceRect := range collisionBounceRects {
				if rectCornersCollideWithPolygon(polygonPath, bounceRect) {
					bounceRectCollision = true
				}
			}

			// check for collisions
			if pathCollision || bounceRectCollision {
				if depth > MAX_RECURSION_DEPTH && rand.Float32() < BACKTRACK_CHANCE {
					backtrackSteps = BACKTRACK_AMOUNT
				}
				// remove polygon path
				polygonPaths = polygonPaths[:polygonPathsStartIndex]

				// return empty bounces
				return []Bounce{}
			}
		}

		// NO COLLISIONS FOUND

		bounceDirPriority := prevBounceDirPriority

		if rand.Float32() < CHANGE_DIR_CHANCE {
			bounceDirPriority[0], bounceDirPriority[1] = bounceDirPriority[1], bounceDirPriority[0]
		}

		// randomly choose whether it is a "horizontal" or "vertical" bounce
		for _, dir := range bounceDirPriority {
			// make square bounce in the direction
			square.InvertDirection(dir)
			square.position = snappedPos

			// create bounce
			bounce := NewBounce(
				len(bounces),
				noteTimeSec,
				snappedPos,
				square.direction,
				dir,
				square.speed,
			)

			// check collision with final bounce rect
			if len(noteOnTimestamps) == 1 {
				collisionBounceRect := bounce.ToCollisionRect()

				for _, pp := range polygonPaths {
					if rectCornersCollideWithPolygon(pp, collisionBounceRect) {
						// remove polygon path
						polygonPaths = polygonPaths[:polygonPathsStartIndex]

						// return empty bounces
						return []Bounce{}
					}
				}
			}

			// save the bounce + safe area
			safeAreas = append(safeAreas, mergeRect(prevSquareRect, square.ToRectangle()))
			bounces = append(bounces, *bounce)

			// make a copy of bounces
			bouncesCopy := make([]Bounce, len(bounces))
			copy(bouncesCopy, bounces)

			// recursive call
			extendedBounces := recursiveGenerate(square, noteOnTimestamps[1:], depth+1, bouncesCopy, noteTimeSec, bounceDirPriority)

			if len(extendedBounces) > 0 {
				return extendedBounces
			}

			// NO PATH FOUND

			// invert direction
			square.InvertDirection(dir)

			// remove the bounce + safe area
			safeAreas = safeAreas[:len(safeAreas)-1]
			bounces = bounces[:len(bounces)-1]

			// remove the path
			if backtrackSteps > 0 {
				backtrackSteps--

				// remove polygon path
				polygonPaths = polygonPaths[:polygonPathsStartIndex]

				// return empty bounces
				return []Bounce{}
			}
		}

		// remove polygon path
		polygonPaths = polygonPaths[:polygonPathsStartIndex]

		// return empty bounces
		return []Bounce{}
	}

	square := NewSquare(rl.NewVector2(0, 0), rl.NewVector2(1, 1), float32(squareSpeed))

	bouncesTemp := recursiveGenerate(square, noteOnTimestamps, 0, []Bounce{}, 0.0, [2]BounceDirection{VerticalBounce, HorizontalBounce})
	if len(bouncesTemp) < 1 {
		return Map{}, errors.New("no path found")
	}
	m.bounces = bouncesTemp

	m.polygonPaths = polygonPaths
	m.safeAreas = mergeOverlappingRects(safeAreas)

	for i := range m.bounces {
		b := &m.bounces[i]

		bounceRect := b.ToRect()

		if rectIsFloating(bounceRect, m.safeAreas) {
			m.floatingBounceRects = append(m.floatingBounceRects, bounceRect)
			b.isFloating = true
		} else {
			m.connectedBounceRects = append(m.connectedBounceRects, bounceRect)

			reachableCells := findReachableCells(bounceRect, m.safeAreas, 75)
			b.reachableCells = reachableCells
		}
	}

	// Post-process to adjust each bounce's speed based on next bounce's position and time
	for i := range len(m.bounces) - 1 {
		current := &m.bounces[i]
		next := m.bounces[i+1]
		deltaTime := next.timeSec - current.timeSec

		var distance float32

		if current.bounceDirection == VerticalBounce {
			distance = float32(math.Abs(float64(next.position.X - current.position.X)))
		} else {
			distance = float32(math.Abs(float64(next.position.Y - current.position.Y)))
		}

		if deltaTime > 0 {
			current.nextSpeed = distance / float32(deltaTime)
		} else {
			current.nextSpeed = float32(squareSpeed)
		}
	}

	return m, nil
}

func rectIsFloating(rect rl.Rectangle, safeAreas []rl.Rectangle) bool {
	for _, r := range safeAreas {
		if rl.CheckCollisionRecs(rect, r) {
			return true
		}
	}

	return false
}

func snapCoordinate(coord float32, cellSize, subCellSize float32) float32 {
	// Which cell are we in?
	cellIndex := float32(math.Floor(float64(coord) / float64(cellSize)))

	// Local coordinate within that cell
	localCoord := coord - cellIndex*cellSize

	// Define an epsilon threshold for deciding "close enough" to a main grid line
	epsilon := float32(0.01)

	// Check if localCoord is near the left/main line (0)
	if rl.CheckCollisionCircleRec(
		rl.NewVector2(localCoord, 0),
		epsilon,
		rl.NewRectangle(0, 0, epsilon, epsilon),
	) {
		localCoord = 0
	} else if localCoord < epsilon {
		// simpler numeric check if within epsilon of 0
		localCoord = 0
	} else if float32(math.Abs(float64(localCoord-cellSize))) < epsilon {
		// near the right boundary, i.e. cellSize
		localCoord = cellSize
	} else {
		// Otherwise, round to nearest multiple of subCellSize
		subIndex := float32(math.Round(float64(localCoord / subCellSize)))
		localCoord = subIndex * subCellSize

		// Also clamp if it somehow overshoots the cell boundary
		if localCoord > cellSize {
			localCoord = cellSize
		}
	}

	return cellIndex*cellSize + localCoord
}

// snapPosition snaps an (x, y) to main grid lines or sub-grid lines
func snapPosition(pos rl.Vector2, cellSize, subCellSize float32) rl.Vector2 {
	return rl.Vector2{
		X: snapCoordinate(pos.X, cellSize, subCellSize),
		Y: snapCoordinate(pos.Y, cellSize, subCellSize),
	}
}
