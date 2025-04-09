package sim

import (
	"math"

	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Polygon []rl.Vector2

// -----------------------------------
// Vertical line vs Polygon
// -----------------------------------
func getPolygonIntersectionsWithVerticalLine(x float32, polygon Polygon) []float32 {
	var intersections []float32
	n := len(polygon)
	if n < 2 {
		return intersections
	}

	for i := 0; i < n; i++ {
		p1 := polygon[i]
		p2 := polygon[(i+1)%n]

		// Check if x is between p1.X and p2.X
		if (p1.X <= x && p2.X >= x) || (p2.X <= x && p1.X >= x) {
			dx := p2.X - p1.X
			// Avoid division by zero or near-zero
			if math.Abs(float64(dx)) < 1e-6 {
				continue
			}
			t := (x - p1.X) / dx
			y := p1.Y + t*(p2.Y-p1.Y)
			intersections = append(intersections, y)
		}
	}
	return intersections
}

func getVerticalSpansInsidePolygon(x, minY, maxY float32, polygon Polygon) []Interval {
	ys := getPolygonIntersectionsWithVerticalLine(x, polygon)
	if len(ys) < 2 {
		return nil
	}

	slices.Sort(ys)

	var inside []Interval
	// Even-odd rule pairing
	for i := 0; i < len(ys)-1; i += 2 {
		start := ys[i]
		end := ys[i+1]
		if start > end {
			start, end = end, start
		}
		// Clip to [minY, maxY]
		if end < minY || start > maxY {
			continue
		}
		if start < minY {
			start = minY
		}
		if end > maxY {
			end = maxY
		}
		inside = append(inside, Interval{start: start, end: end})
	}
	return inside
}

func drawVerticalLineOutsidePolygons(x, startY, endY float32, polygons []Polygon, color rl.Color) {
	segments := []Interval{
		{start: startY, end: endY},
	}
	// Subtract polygon intervals
	for _, poly := range polygons {
		insideSpans := getVerticalSpansInsidePolygon(x, startY, endY, poly)
		for _, span := range insideSpans {
			segments = subtractInterval(segments, span.start, span.end)
		}
	}

	// Draw remaining segments
	for _, seg := range segments {
		if seg.end > seg.start {
			rl.DrawLine(
				int32(x),
				int32(seg.start),
				int32(x),
				int32(seg.end),
				color,
			)
		}
	}
}

// -----------------------------------
// Horizontal line vs Polygon
// -----------------------------------
func getPolygonIntersectionsWithHorizontalLine(y float32, polygon Polygon) []float32 {
	var intersections []float32
	n := len(polygon)
	if n < 2 {
		return intersections
	}

	for i := range n {
		p1 := polygon[i]
		p2 := polygon[(i+1)%n]

		if (p1.Y <= y && p2.Y >= y) || (p2.Y <= y && p1.Y >= y) {
			dy := p2.Y - p1.Y
			if math.Abs(float64(dy)) < 1e-6 {
				continue
			}
			t := (y - p1.Y) / dy
			x := p1.X + t*(p2.X-p1.X)
			intersections = append(intersections, x)
		}
	}
	return intersections
}

func getHorizontalSpansInsidePolygon(y, minX, maxX float32, polygon Polygon) []Interval {
	xs := getPolygonIntersectionsWithHorizontalLine(y, polygon)
	if len(xs) < 2 {
		return nil
	}

	slices.Sort(xs)

	var inside []Interval
	// Even-odd pairing
	for i := 0; i < len(xs)-1; i += 2 {
		start := xs[i]
		end := xs[i+1]
		if start > end {
			start, end = end, start
		}
		if end < minX || start > maxX {
			continue
		}
		if start < minX {
			start = minX
		}
		if end > maxX {
			end = maxX
		}
		inside = append(inside, Interval{start: start, end: end})
	}
	return inside
}

func drawHorizontalLineOutsidePolygons(y, startX, endX float32, polygons []Polygon, color rl.Color) {
	segments := []Interval{
		{start: startX, end: endX},
	}
	for _, poly := range polygons {
		insideSpans := getHorizontalSpansInsidePolygon(y, startX, endX, poly)
		for _, span := range insideSpans {
			segments = subtractInterval(segments, span.start, span.end)
		}
	}

	for _, seg := range segments {
		if seg.end > seg.start {
			rl.DrawLine(
				int32(seg.start),
				int32(y),
				int32(seg.end),
				int32(y),
				color,
			)
		}
	}
}

func drawGridOutsidePolygons(
	startX, endX, startY, endY float32,
	cellSize int32,
	color rl.Color,
	polygons []Polygon,
) {
	// Draw vertical lines
	for x := startX; x <= endX; x += float32(cellSize) {
		drawVerticalLineOutsidePolygons(x, startY, endY, polygons, color)
	}
	// Draw horizontal lines
	for y := startY; y <= endY; y += float32(cellSize) {
		drawHorizontalLineOutsidePolygons(y, startX, endX, polygons, color)
	}
}

func PixelsToCellSet(pixels []rl.Vector2, cellSize int) map[[2]int]bool {
	cells := make(map[[2]int]bool)
	for _, v := range pixels {
		cx := int(v.X) / cellSize
		cy := int(v.Y) / cellSize
		cells[[2]int{cx, cy}] = true
	}
	return cells
}

type segment struct {
	p1, p2 rl.Vector2
}

// CellsToSinglePolygon converts a set of occupied cells into a single boundary polygon.
// It assumes *one connected region* with no holes. If your shape can be disconnected
// or have holes, you'd need more complex logic.
func CellsToSinglePolygon(cells map[[2]int]bool, cellSize int) []rl.Vector2 {
	// 1) Collect boundary edges
	var boundarySegments []segment

	for c := range cells {
		cx, cy := c[0], c[1]

		// For each of the 4 edges, check neighbor occupancy
		// We'll produce an edge if that neighbor is not occupied.
		// E.g. top neighbor is (cx, cy-1)

		// Top edge:
		if !cells[[2]int{cx, cy - 1}] {
			p1 := rl.NewVector2(float32(cx*cellSize), float32(cy*cellSize))
			p2 := rl.NewVector2(float32((cx+1)*cellSize), float32(cy*cellSize))
			boundarySegments = append(boundarySegments, segment{p1, p2})
		}

		// Right edge:
		if !cells[[2]int{cx + 1, cy}] {
			p1 := rl.NewVector2(float32((cx+1)*cellSize), float32(cy*cellSize))
			p2 := rl.NewVector2(float32((cx+1)*cellSize), float32((cy+1)*cellSize))
			boundarySegments = append(boundarySegments, segment{p1, p2})
		}

		// Bottom edge:
		if !cells[[2]int{cx, cy + 1}] {
			p1 := rl.NewVector2(float32(cx*cellSize), float32((cy+1)*cellSize))
			p2 := rl.NewVector2(float32((cx+1)*cellSize), float32((cy+1)*cellSize))
			boundarySegments = append(boundarySegments, segment{p1, p2})
		}

		// Left edge:
		if !cells[[2]int{cx - 1, cy}] {
			p1 := rl.NewVector2(float32(cx*cellSize), float32(cy*cellSize))
			p2 := rl.NewVector2(float32(cx*cellSize), float32((cy+1)*cellSize))
			boundarySegments = append(boundarySegments, segment{p1, p2})
		}
	}

	// 2) Chain edges into a single loop of vertices
	polygon := chainEdgesIntoOneLoop(boundarySegments)
	return polygon
}

// chainEdgesIntoOneLoop tries to connect a list of segments into a single loop of vertices.
// We assume there's exactly one closed boundary (no holes, no disconnected parts).
func chainEdgesIntoOneLoop(segs []segment) []rl.Vector2 {
	// Build adjacency: each point -> list of connected points
	adjacency := make(map[rl.Vector2][]rl.Vector2)
	for _, s := range segs {
		adjacency[s.p1] = append(adjacency[s.p1], s.p2)
		adjacency[s.p2] = append(adjacency[s.p2], s.p1)
	}

	// Find a "start" point. For simplicity, pick the first segment’s p1.
	if len(segs) == 0 {
		return nil
	}
	start := segs[0].p1

	var polygon []rl.Vector2
	polygon = append(polygon, start)

	current := start
	var prev rl.Vector2 // zero for now

	for {
		nextPoints := adjacency[current]
		// We have up to 2 possible next points (since it's a boundary).
		// One of them is "prev", the other is the new direction
		var next rl.Vector2
		found := false

		for _, candidate := range nextPoints {
			if !almostEqual(candidate, prev) {
				next = candidate
				found = true
				break
			}
		}

		if !found {
			// No way forward -> break
			break
		}

		polygon = append(polygon, next)

		// Move along
		prev = current
		current = next

		if almostEqual(current, start) {
			// We are back at the start -> closed loop
			break
		}
	}

	return polygon
}

// For floating precision, you often want a small epsilon compare
func almostEqual(a, b rl.Vector2) bool {
	const eps = 0.0001
	return (rl.Vector2Distance(a, b) < eps)
}

func drawVerticalLineOutsideRects(x, startY, endY float32, rects []rl.Rectangle, color rl.Color) {
	segments := []Interval{
		{start: startY, end: endY},
	}

	for _, rect := range rects {
		// Check if x is within the rect's X bounds (exclusive of rect.X + rect.Width)
		if x > rect.X && x < rect.X+rect.Width {
			skipY1 := rect.Y
			skipY2 := rect.Y + rect.Height
			segments = subtractInterval(segments, skipY1, skipY2)
		}
	}

	for _, seg := range segments {
		if seg.end > seg.start {
			rl.DrawLine(int32(x), int32(seg.start), int32(x), int32(seg.end), color)
		}
	}
}

func drawHorizontalLineOutsideRects(y, startX, endX float32, rects []rl.Rectangle, color rl.Color) {
	segments := []Interval{
		{start: startX, end: endX},
	}

	for _, rect := range rects {
		// Check if y is within the rect's Y bounds (exclusive of rect.Y + rect.Height)
		if y > rect.Y && y < rect.Y+rect.Height {
			skipX1 := rect.X
			skipX2 := rect.X + rect.Width
			segments = subtractInterval(segments, skipX1, skipX2)
		}
	}

	for _, seg := range segments {
		if seg.end > seg.start {
			rl.DrawLine(int32(seg.start), int32(y), int32(seg.end), int32(y), color)
		}
	}
}

func drawGridOutsideRects(startX, endX, startY, endY float32, cellSize int, color rl.Color, rects []rl.Rectangle) {
	// draw vertical lines skipping rects
	for x := startX; x <= endX; x += float32(cellSize) {
		drawVerticalLineOutsideRects(x, startY, endY, rects, color)
	}

	// draw horizontal lines skipping rects
	for y := startY; y <= endY; y += float32(cellSize) {
		drawHorizontalLineOutsideRects(y, startX, endX, rects, color)
	}
}

func drawVerticalLineInsideRects(x, startY, endY float32, rects []rl.Rectangle, color rl.Color) {
	// We'll collect Intervals that lie within any rect.
	var segments []Interval

	// For each rect, if `x` is inside that rect's [X, X+Width],
	// add [rect.Y, rect.Y+rect.Height] to the segments, clipped to [startY, endY].
	for _, rect := range rects {
		if x >= rect.X && x <= rect.X+rect.Width {
			top := rect.Y
			bot := rect.Y + rect.Height

			// fmt.Printf("%v\n", rect.Height/SUB_CELL_SIZE)

			// Clip to [startY, endY] so we don't draw beyond the camera region
			if bot < startY || top > endY {
				continue
			}

			if top < startY {
				top = startY
			}
			if bot > endY {
				bot = endY
			}
			segments = addInterval(segments, Interval{start: top, end: bot})
		}
	}

	// Draw each final merged segment
	for _, seg := range segments {
		if seg.end > seg.start {
			rl.DrawLine(int32(x), int32(seg.start), int32(x), int32(seg.end), color)
		}
	}
}

func drawHorizontalLineInsideRects(y, startX, endX float32, rects []rl.Rectangle, color rl.Color) {
	var segments []Interval

	// For each bounce rect, if y is inside [rect.Y, rect.Y+rect.Height],
	// add [rect.X, rect.X+rect.Width] to the segments, clipped to [startX, endX].
	for _, rect := range rects {
		if y >= rect.Y && y <= rect.Y+rect.Height {
			left := rect.X
			right := rect.X + rect.Width

			// Clip to [startX, endX] so we don't draw beyond the camera region
			if right < startX || left > endX {
				continue
			}

			if left < startX {
				left = startX
			}
			if right > endX {
				right = endX
			}
			segments = addInterval(segments, Interval{start: left, end: right})
		}
	}

	for _, seg := range segments {
		if seg.end > seg.start {
			rl.DrawLine(int32(seg.start), int32(y), int32(seg.end), int32(y), color)
		}
	}
}

func drawGridInsideRects(startX, endX, startY, endY float32, cellSize int, color rl.Color, rects []rl.Rectangle) {
	// Draw vertical sub‐grid lines only inside bounce rects
	for x := startX; x <= endX; x += float32(cellSize) {
		drawVerticalLineInsideRects(x, startY, endY, rects, color)
	}

	// Draw horizontal sub‐grid lines only inside bounce rects
	for y := startY; y <= endY; y += float32(cellSize) {
		drawHorizontalLineInsideRects(y, startX, endX, rects, color)
	}
}

//------------------------------------------------------------------------------
// Drawing vertical/horizontal lines using INCLUDE minus EXCLUDE
//------------------------------------------------------------------------------

// drawVerticalLineIncludeExclude draws a vertical line (at x) for intervals that lie
// in the union of includeRects MINUS the union of excludeRects, clipped to [startY, endY].
func drawVerticalLineIncludeExclude(
	x, startY, endY float32,
	includeRects, excludeRects []rl.Rectangle,
	color rl.Color,
) {
	// 1. Gather merged “include” intervals
	includeSegments := gatherVerticalIntervals(x, startY, endY, includeRects)

	// 2. Gather merged “exclude” intervals
	excludeSegments := gatherVerticalIntervals(x, startY, endY, excludeRects)

	// 3. Subtract exclude from include
	finalSegments := includeSegments
	for _, exc := range excludeSegments {
		finalSegments = subtractInterval(finalSegments, exc.start, exc.end)
	}

	// 4. Draw final segments
	for _, seg := range finalSegments {
		if seg.end > seg.start {
			rl.DrawLine(int32(x), int32(seg.start), int32(x), int32(seg.end), color)
		}
	}
}

// drawHorizontalLineIncludeExclude draws a horizontal line (at y) for intervals that lie
// in the union of includeRects MINUS the union of excludeRects, clipped to [startX, endX].
func drawHorizontalLineIncludeExclude(
	y, startX, endX float32,
	includeRects, excludeRects []rl.Rectangle,
	color rl.Color,
) {
	// 1. Gather merged “include” intervals
	includeSegments := gatherHorizontalIntervals(y, startX, endX, includeRects)

	// 2. Gather merged “exclude” intervals
	excludeSegments := gatherHorizontalIntervals(y, startX, endX, excludeRects)

	// 3. Subtract exclude from include
	finalSegments := includeSegments
	for _, exc := range excludeSegments {
		finalSegments = subtractInterval(finalSegments, exc.start, exc.end)
	}

	// 4. Draw final segments
	for _, seg := range finalSegments {
		if seg.end > seg.start {
			rl.DrawLine(int32(seg.start), int32(y), int32(seg.end), int32(y), color)
		}
	}
}

//------------------------------------------------------------------------------
// Main "drawGridIncludeExcludeRects" function
//------------------------------------------------------------------------------

// drawGridIncludeExcludeRects draws a grid (vertical/horizontal lines at `cellSize` spacing)
// only in areas that are inside the union of includeRects BUT outside any excludeRect.
func drawGridIncludeExcludeRects(startX, endX, startY, endY float32, cellSize int, color rl.Color, includeRects, excludeRects []rl.Rectangle) {
	// Draw vertical grid lines
	for x := startX; x <= endX; x += float32(cellSize) {
		drawVerticalLineIncludeExclude(x, startY, endY, includeRects, excludeRects, color)
	}

	// Draw horizontal grid lines
	for y := startY; y <= endY; y += float32(cellSize) {
		drawHorizontalLineIncludeExclude(y, startX, endX, includeRects, excludeRects, color)
	}
}
