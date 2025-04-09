package sim

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var colorWaves []*ColorWave

type ColorWave struct {
	rect    rl.Rectangle
	maxSize int

	startTime float64
	bounceIdx int

	expandDirection rl.Vector2
	isExpanding     bool

	color rl.Color
	speed float32
}

func NewColorWave(center rl.Vector2, initialSize float32, maxSize int, startTime float64, bounceIdx int, expandDirection rl.Vector2, color rl.Color, speed float32) ColorWave {
	return ColorWave{
		rect:    rl.NewRectangle(center.X-initialSize/2, center.Y-initialSize/2, initialSize, initialSize),
		maxSize: maxSize,

		startTime: startTime,
		bounceIdx: bounceIdx,

		expandDirection: expandDirection,
		isExpanding:     true,

		color: color,
		speed: speed,
	}
}

func (cw ColorWave) IsExpanding() bool {
	return cw.isExpanding
}

func (cw ColorWave) GetRect() rl.Rectangle {
	return cw.rect
}

func (cw *ColorWave) Update(dt float32) {
	// If the wave isn’t expanding anymore, do nothing
	if cw.rect.Width >= float32(cw.maxSize) {
		if cw.isExpanding {
			// snappedPos := snapPosition(rl.NewVector2(cw.rect.X, cw.rect.Y), CELL_SIZE, CELL_SIZE)

			// cw.rect.Height = float32(cw.maxSize)
			// cw.rect.Width = float32(cw.maxSize)

			cw.isExpanding = false
		}
		return
	}

	// expanding on x axis
	// if cw.expandDirection.X != 0 {
	// 	if cw.rect.Height >= float32(cw.maxSize) {
	// 		cw.rect.Height = float32(cw.maxSize)
	// 		cw.isExpanding = false
	// 		return
	// 	}
	// } else if cw.rect.Width >= float32(cw.maxSize) {
	// 	cw.rect.Width = float32(cw.maxSize)
	// 	cw.isExpanding = false
	// 	return
	// }

	cw.rect.X -= cw.speed * dt
	cw.rect.Y -= cw.speed * dt
	cw.rect.Width += cw.speed * 2 * dt
	cw.rect.Height += cw.speed * 2 * dt

	// TODO snap to grid
	// TODO make it directional exapnsion (where the square was heading when it hit the wall)

	// move and expand based on expandDirection, for example (1, 0) is right, (-1, 0) is left, (0, 1) is down, (0, -1) is up
	// if expanding to the right, rect should not move to the left, but expand to the right (and up and down)
	// if expanding to the left, rect should not move to the right, but expand to the left (and up and down)
	// if expanding down, rect should not move up, but expand down (and left and right)
	// if expanding up, rect should not move down, but expand up (and left and right)
	// switch cw.expandDirection {
	// case rl.Vector2{X: 1, Y: 0}:
	// 	cw.rect.Height += cw.speed * 2 * dt
	// 	cw.rect.Width += cw.speed * dt
	// 	cw.rect.Y -= cw.speed * dt

	// case rl.Vector2{X: -1, Y: 0}:
	// 	cw.rect.Height += cw.speed * 2 * dt
	// 	cw.rect.Width += cw.speed * dt
	// 	cw.rect.Y -= cw.speed * dt
	// 	cw.rect.X -= cw.speed * dt

	// case rl.Vector2{X: 0, Y: 1}:
	// 	cw.rect.Width += cw.speed * 2 * dt
	// 	cw.rect.Height += cw.speed * dt
	// 	cw.rect.X -= cw.speed * dt

	// case rl.Vector2{X: 0, Y: -1}:
	// 	cw.rect.Width += cw.speed * 2 * dt
	// 	cw.rect.Height += cw.speed * dt
	// 	cw.rect.X -= cw.speed * dt
	// 	cw.rect.Y -= cw.speed * dt
	// }
}
