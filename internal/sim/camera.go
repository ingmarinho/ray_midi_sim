package sim

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TODO make camera a struct

func followCameraSmooth(camera *rl.Camera2D, target rl.Vector2, dt float32) {
	// TODO make these constants
	minSpeed := 30        // minimum speed of the camera
	minEffectLength := 10 // minimum distance to start the effect
	fractionSpeed := 2.5  // speed of the camera effect

	camera.Offset = WINDOW_CENTER_VECTOR
	diff := rl.Vector2Subtract(target, camera.Target)
	length := rl.Vector2Length(diff)

	if length > float32(minEffectLength) {
		speed := math.Max(float64(minSpeed), fractionSpeed*float64(length))
		camera.Target = rl.Vector2Add(camera.Target, rl.Vector2Scale(diff, float32(speed)*dt/length))
	}
}

func GetCameraRect(camera rl.Camera2D, screenWidth, screenHeight int32) rl.Rectangle {
	halfW := float32(screenWidth) * 0.5 / camera.Zoom
	halfH := float32(screenHeight) * 0.5 / camera.Zoom

	left := camera.Target.X - halfW
	right := camera.Target.X + halfW
	top := camera.Target.Y - halfH
	bottom := camera.Target.Y + halfH

	return rl.NewRectangle(left, top, right-left, bottom-top)
}

func GetCameraBoundaries(cameraRect rl.Rectangle, cellSize int32) (startX, endX, startY, endY float32) {
	// snap the boundaries to multiples of cellSize
	startX = float32(math.Floor(float64(cameraRect.X)/float64(cellSize))) * float32(cellSize)
	endX = float32(math.Ceil(float64(cameraRect.X+cameraRect.Width)/float64(cellSize))) * float32(cellSize)
	startY = float32(math.Floor(float64(cameraRect.Y)/float64(cellSize))) * float32(cellSize)
	endY = float32(math.Ceil(float64(cameraRect.Y+cameraRect.Height)/float64(cellSize))) * float32(cellSize)

	// add margin of 1 cell around the camera
	margin := float32(cellSize)
	startX -= margin
	endX += margin
	startY -= margin
	endY += margin

	return startX, endX, startY, endY
}
