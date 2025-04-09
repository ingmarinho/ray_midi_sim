package sim

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Square struct {
	position  rl.Vector2
	direction rl.Vector2
	speed     float32

	// Animation fields
	bounceAnimTimer    float32 // how long the current bounce animation has been running
	bounceAnimDuration float32 // total time the bounce animation should last

	// Track if the last bounce was horizontal or vertical
	bounceDirection BounceDirection
}

func (s *Square) GetPosition() rl.Vector2 {
	return s.position
}

func (s *Square) InvertDirection(bounceDirection BounceDirection) {
	switch bounceDirection {
	case HorizontalBounce:
		s.direction.X *= -1
	case VerticalBounce:
		s.direction.Y *= -1
	}
}

func (s *Square) Bounce(bounce Bounce, bounceIdx int) {
	prevDirection := s.direction

	s.position = bounce.position
	s.direction = bounce.nextDirection
	s.speed = bounce.nextSpeed

	// Increase this duration to slow down the bounce animation
	s.bounceAnimTimer = 0
	s.bounceAnimDuration = 0.5

	// Record the bounce direction to determine squash & stretch orientation
	s.bounceDirection = bounce.bounceDirection

	if !bounce.isFloating && false {
		// only left, right up and down, no diagonal, also check bounce direction to determine the direction of the expansion so (1, 0) is right, (-1, 0) is left, (0, 1) is down, (0, -1) is up
		// bounceRect := bounce.ToRect()
		// bounceRectCenter := rl.NewVector2(bounceRect.X+bounceRect.Width/2, bounceRect.Y+bounceRect.Height/2)
		// squareCenter := rl.NewVector2(s.position.X+SQUARE_SIZE/2, s.position.Y+SQUARE_SIZE/2)
		// isEven := max(BOUNCE_RECT_HEIGHT, BOUNCE_RECT_WIDTH)/CELL_SIZE%2 == 0

		expandDirection := rl.Vector2{X: 0, Y: 0}
		switch s.bounceDirection {
		case HorizontalBounce:
			expandDirection.X = prevDirection.X
			// if !isEven {
			// 	bounceRectCenter.Y -= CELL_SIZE / 2
			// }
		case VerticalBounce:
			expandDirection.Y = prevDirection.Y
			// if !isEven {
			// 	bounceRectCenter.X -= CELL_SIZE / 2
			// }
		}

		// switch expandDirection {
		// // right
		// case rl.NewVector2(1, 0):
		// 	bounceRectCenter.X -= CELL_SIZE / 2
		// 	// left
		// case rl.NewVector2(-1, 0):
		// 	bounceRectCenter.X += CELL_SIZE / 2
		// 	// down
		// case rl.NewVector2(0, 1):
		// 	bounceRectCenter.Y -= CELL_SIZE / 2
		// 	// up
		// case rl.NewVector2(0, -1):
		// 	bounceRectCenter.Y += CELL_SIZE / 2
		// }

		ALL_COLORS := []rl.Color{
			rl.Pink,
		}
		color := ALL_COLORS[rand.Intn(len(ALL_COLORS))]

		cw := NewColorWave(s.position, CELL_SIZE, 700, rl.GetTime(), bounceIdx, expandDirection, color, 200)

		colorWaves = append(colorWaves, &cw)
	}
}

func (s *Square) Update(dt float32) {
	// Regular movement
	s.position.X += s.direction.X * s.speed * dt
	s.position.Y += s.direction.Y * s.speed * dt

	// Update the bounce animation timer
	if s.bounceAnimTimer < s.bounceAnimDuration {
		s.bounceAnimTimer += dt
		if s.bounceAnimTimer > s.bounceAnimDuration {
			s.bounceAnimTimer = s.bounceAnimDuration
		}
	}
}

func (s *Square) Draw() {
	// Default scales
	scaleX := float32(1.0)
	scaleY := float32(1.0)

	// If we’re within the bounce animation window, apply squash & stretch
	if s.bounceAnimTimer < s.bounceAnimDuration {
		// progress goes from 0.0 (start) to 1.0 (end)
		progress := s.bounceAnimTimer / s.bounceAnimDuration

		// TODO make these constants
		amplitude := float32(0.4) // "strength" of the stretch
		frequency := float64(1.5) // how many oscillations per second
		damping := float64(2.5)   // how quickly it decays

		// Example damped sine wave
		osc := float32(math.Exp(-damping*float64(progress))) *
			float32(math.Sin(2.0*math.Pi*frequency*float64(progress)))

		// Switch scaling based on bounce direction
		switch s.bounceDirection {
		case VerticalBounce:
			scaleX = 1.0 + amplitude*osc
			scaleY = 1.0 - amplitude*osc

		case HorizontalBounce:
			scaleX = 1.0 - amplitude*osc
			scaleY = 1.0 + amplitude*osc
		}
	}

	// Calculate how large the rectangle is after scaling
	scaledWidth := SQUARE_SIZE * scaleX
	scaledHeight := SQUARE_SIZE * scaleY

	// Offset so the scaled rectangle is still centered at s.position
	drawPos := rl.NewVector2(
		s.position.X+(SQUARE_SIZE/2-scaledWidth/2),
		s.position.Y+(SQUARE_SIZE/2-scaledHeight/2),
	)

	sizeVector := rl.NewVector2(scaledWidth, scaledHeight)

	outlineThickness := float32(3.0)

	// Draw the square's outline
	rl.DrawRectangleV(drawPos, rl.NewVector2(scaledWidth, scaledHeight), rl.White)

	// Draw the square
	rl.DrawRectangleV(
		rl.NewVector2(drawPos.X+outlineThickness, drawPos.Y+outlineThickness),
		rl.NewVector2(sizeVector.X-outlineThickness*2, sizeVector.Y-outlineThickness*2),
		rl.Red,
	)
}

func (s Square) ToRectangle() rl.Rectangle {
	return rl.NewRectangle(s.position.X, s.position.Y, SQUARE_SIZE, SQUARE_SIZE)
}

func NewSquare(position rl.Vector2, direction rl.Vector2, speed float32) Square {
	return Square{
		position:  position,
		direction: direction,
		speed:     speed,
	}
}
