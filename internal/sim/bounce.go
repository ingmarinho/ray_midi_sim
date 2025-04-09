package sim

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type BounceDirection int

const (
	HorizontalBounce BounceDirection = iota
	VerticalBounce
)

type Bounce struct {
	id              int
	timeSec         float64
	position        rl.Vector2
	nextDirection   rl.Vector2
	bounceDirection BounceDirection
	nextSpeed       float32

	reachableCells []Cell
	isFloating     bool
}

func (b Bounce) IsFloating() bool {
	return b.isFloating
}

func (b *Bounce) Draw() {
	rect := b.ToRect()
	rl.DrawRectangleRec(rect, rl.Blue)
	rl.DrawRectangleLinesEx(rect, 1, rl.Black)
}

func (b Bounce) ToCollisionRect() rl.Rectangle {
	bounceRect := b.ToRect()

	const OFFSET float32 = 1

	switch b.bounceDirection {
	case HorizontalBounce:
		if b.nextDirection.X == 1 {
			// left wall
			bounceRect.X -= OFFSET
		} else if b.nextDirection.X == -1 {
			// right wall
			bounceRect.X += OFFSET
		}
	case VerticalBounce:
		if b.nextDirection.Y == 1 {
			// top wall
			bounceRect.Y -= OFFSET
		} else if b.nextDirection.Y == -1 {
			// bottom wall
			bounceRect.Y += OFFSET
		}
	}

	return bounceRect
}

func (b Bounce) ToRect() rl.Rectangle {
	switch b.bounceDirection {
	case HorizontalBounce:
		if b.nextDirection.X == 1 {
			// left wall
			return rl.NewRectangle(
				b.position.X-BOUNCE_RECT_WIDTH,
				b.position.Y+SQUARE_SIZE/2-BOUNCE_RECT_HEIGHT/2,
				BOUNCE_RECT_WIDTH,
				BOUNCE_RECT_HEIGHT,
			)

		} else if b.nextDirection.X == -1 {
			// right wall
			return rl.NewRectangle(
				b.position.X+SQUARE_SIZE,
				b.position.Y+SQUARE_SIZE/2-BOUNCE_RECT_HEIGHT/2,
				BOUNCE_RECT_WIDTH,
				BOUNCE_RECT_HEIGHT,
			)
		}
	case VerticalBounce:
		if b.nextDirection.Y == 1 {
			// top wall
			return rl.NewRectangle(
				b.position.X+SQUARE_SIZE/2-BOUNCE_RECT_HEIGHT/2,
				b.position.Y-BOUNCE_RECT_WIDTH,
				BOUNCE_RECT_HEIGHT,
				BOUNCE_RECT_WIDTH,
			)

		} else if b.nextDirection.Y == -1 {
			// bottom wall
			return rl.NewRectangle(
				b.position.X+SQUARE_SIZE/2-BOUNCE_RECT_HEIGHT/2,
				b.position.Y+SQUARE_SIZE,
				BOUNCE_RECT_HEIGHT,
				BOUNCE_RECT_WIDTH,
			)
		}
	}

	// unreachable
	return rl.NewRectangle(0, 0, 0, 0)
}

func NewBounce(id int, timeSec float64, position rl.Vector2, nextDirection rl.Vector2, bounceDirection BounceDirection, nextSpeed float32) *Bounce {
	return &Bounce{id: id, timeSec: timeSec, position: position, nextDirection: nextDirection, bounceDirection: bounceDirection, nextSpeed: nextSpeed}
}
