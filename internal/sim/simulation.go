package sim

import (
	"ray_midi_sim/internal/midi"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Simulation struct {
	// paths
	midPath string
	wavPath string

	// requires initialisation
	generatedMap     Map
	midi             midi.Midi
	music            rl.Music
	noteOnTimestamps []float64
	square           Square
	camera           rl.Camera2D

	// simulation state
	squareMoving       bool
	startTimeSec       float64
	currentTimeSec     float64
	bounceIdx          int
	floatingBounceIdx  int
	connectedBounceIdx int
}

func New(midPath, wavPath string) Simulation {
	return Simulation{
		midPath: midPath,
		wavPath: wavPath,
	}
}

func (s *Simulation) Init() error {
	// initialise raylib stuff
	rl.InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "RAY MIDI SIM")
	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagVsyncHint)
	rl.SetTargetFPS(FPS)
	rl.InitAudioDevice()

	s.music = rl.LoadMusicStream(s.wavPath)

	// initialise midi stuff
	midiTemp, err := midi.New(s.midPath)
	if err != nil {
		return err
	}
	s.midi = midiTemp

	// s.midi.TrimByDuration(5)

	noteOnTimestamps := s.midi.ExtractNoteOnTimestamps()
	noteOnTimestampsFiltered := midi.FilterTimestampsByCloseness(noteOnTimestamps, CLOSENESS_THRESHOLD_MS)
	s.noteOnTimestamps = noteOnTimestampsFiltered

	// generate map
	generatedMapTemp, err := GenerateMap(s.noteOnTimestamps, SQUARE_SPEED)
	if err != nil {
		return err
	}
	s.generatedMap = generatedMapTemp

	// initialise square
	s.square = NewSquare(rl.NewVector2(0, 0), rl.NewVector2(1, 1), SQUARE_SPEED)
	s.camera = rl.NewCamera2D(WINDOW_CENTER_VECTOR, s.square.GetPosition(), 0, 1)

	return nil
}

func (s *Simulation) update() {
	rl.UpdateMusicStream(s.music)

	dt := rl.GetFrameTime()
	s.currentTimeSec = rl.GetTime() - s.startTimeSec
	s.bounceIdx = s.floatingBounceIdx + s.connectedBounceIdx

	// update color waves
	for _, wave := range colorWaves {
		wave.Update(dt)
	}

	// start music and square movement
	if s.currentTimeSec >= 0.0 && !rl.IsMusicStreamPlaying(s.music) {
		rl.PlayMusicStream(s.music)

		s.squareMoving = true
	}

	// handle bounce
	if s.bounceIdx < len(s.generatedMap.bounces) && s.currentTimeSec >= float64(s.generatedMap.bounces[s.bounceIdx].timeSec) {
		currentBounce := s.generatedMap.bounces[s.bounceIdx]

		s.square.Bounce(currentBounce, s.bounceIdx)

		if currentBounce.IsFloating() {
			s.floatingBounceIdx++
		} else {
			s.connectedBounceIdx++
		}

		s.bounceIdx++
	}

	if s.bounceIdx >= len(s.generatedMap.bounces) {
		s.squareMoving = false
	}

	// update square movement
	if s.squareMoving {
		s.square.Update(dt)
	}

	// update camera
	centeredSquarePos := rl.Vector2AddValue(s.square.GetPosition(), SQUARE_SIZE/2)
	followCameraSmooth(&s.camera, centeredSquarePos, dt)

	// zoom in/out with mouse wheel (TEMPORARY FOR TESTING)
	s.camera.Zoom += rl.GetMouseWheelMove() * 0.05
}

func (s *Simulation) draw() {
	cameraRect := GetCameraRect(s.camera, WINDOW_WIDTH, WINDOW_HEIGHT)
	startX, endX, startY, endY := GetCameraBoundaries(cameraRect, CELL_SIZE)

	rl.BeginDrawing()
	{
		// TODO make background color a constant
		// rl.ClearBackground(rl.White)

		// clear background with gradient
		rl.DrawRectangleGradientV(0, 0, WINDOW_WIDTH, WINDOW_HEIGHT, rl.NewColor(124, 0, 1, 255), rl.Black)
		rl.BeginMode2D(s.camera)
		{
			// toBeDrawn := make([]int, 0)

			// // // draw color waves
			// for _, bounce := range s.generatedMap.bounces[:s.bounceIdx] {
			// 	if bounce.IsFloating() {
			// 		continue
			// 	}

			// 	// TODO optimise this by only checking the color wave rect that the bounce is in
			// 	// TODO check center of cell instead of only top left corner (otherwise the expansion will be faster on some cells)
			// 	// TODO general cleanup/optimisation of code
			// 	// TODO improve cell finding algorithm
			// 	// TODO add different colors for different bounces
			// 	// TODO add lines of color within a color wave

			// 	// TODO image pixel art idea on the color wave expansion

			// 	for i, wave := range colorWaves {
			// 		// if wave.bounceIdx == bounce.id {
			// 		// 	// only draw the color wave if it's within the camera view
			// 		// 	// if !rl.CheckCollisionRecs(wave.rect, cameraRect) {
			// 		// 	// 	continue
			// 		// 	// }

			// 		// 	if wave.isExpanding {
			// 		// 		if len(wave.cells) == len(bounce.reachableCells) {
			// 		// 			wave.isExpanding = false
			// 		// 		}

			// 		// 		if !rl.CheckCollisionCircleRec(wave.center, wave.radius, cameraRect) {
			// 		// 			continue
			// 		// 		}

			// 		// 		for _, c := range bounce.reachableCells {
			// 		// 			if rl.CheckCollisionPointCircle(c.pos, wave.center, wave.radius) {
			// 		// 				wave.AddCell(c)
			// 		// 			}
			// 		// 		}

			// 		// 		rl.DrawCircleGradient(int32(wave.center.X), int32(wave.center.Y), wave.radius, rl.Fade(wave.color, 0), rl.Fade(wave.color, 100))
			// 		// 	}

			// 		// 	cellRects := make([]rl.Rectangle, 0)
			// 		// 	for _, c := range wave.cells {
			// 		// 		cellRects = append(cellRects, c.ToRect())
			// 		// 	}

			// 		// 	for _, c := range wave.cells {
			// 		// 		rl.DrawRectangleRec(c.ToRect(), wave.color)
			// 		// 	}
			// 		// 	// drawGridInsideRects(startX, endX, startY, endY, CELL_SIZE, rl.Maroon, cellRects)
			// 		// 	// rl.DrawCircleLinesV(wave.center, wave.radius, rl.Black)

			// 		// 	break
			// 		// }

			// 		if wave.bounceIdx == bounce.id {
			// 			if !rl.CheckCollisionRecs(wave.rect, cameraRect) {
			// 				continue
			// 			}

			// 			// if wave.isExpanding {
			// 			// 	inflatedRect := rl.NewRectangle(wave.rect.X-CELL_SIZE/2, wave.rect.Y-CELL_SIZE/2, wave.rect.Width+CELL_SIZE, wave.rect.Height+CELL_SIZE)
			// 			// 	rl.DrawRectangleRec(inflatedRect, rl.Black)
			// 			// }

			// 			toBeDrawn = append(toBeDrawn, i)

			// 			break
			// 		}
			// 	}

			// 	// 	// for _, i := range toBeDrawn {
			// 	// 	// 	rl.DrawRectangleRec(colorWaves[i].rect, colorWaves[i].color)
			// 	// 	// }

			// 	// 	// drawGridIncludeExcludeRects(startX, endX, startY, endY, CELL_SIZE, rl.Orange, toBeDrawn, s.generatedMap.safeAreas)
			// }

			// for _, i := range toBeDrawn {
			// 	rl.DrawRectangleRec(colorWaves[i].rect, colorWaves[i].color)
			// }

			// for _, safeArea := range s.generatedMap.safeAreas {
			// 	rl.DrawRectangleRec(safeArea, rl.NewColor(140, 140, 140, 255))
			// }

			// for _, i := range toBeDrawn {
			// 	rl.DrawRectangleLinesEx(colorWaves[i].rect, 1, rl.Black)
			// }

			// draw grid
			drawGridOutsideRects(startX, endX, startY, endY, CELL_SIZE, rl.Black, s.generatedMap.safeAreas)

			merged := make([]rl.Rectangle, 0)
			for _, bounce := range s.generatedMap.bounces {
				merged = append(merged, bounce.ToRect())
			}

			drawGridInsideRects(startX, endX, startY, endY, CELL_SIZE, rl.Black, merged[s.bounceIdx:])
			drawGridInsideRects(startX, endX, startY, endY, CELL_SIZE, rl.Red, merged[:s.bounceIdx])

			// draw floating bounces
			// drawGridInsideRects(startX, endX, startY, endY, CELL_SIZE, rl.White, s.generatedMap.floatingBounceRects[s.floatingBounceIdx:])
			// drawGridInsideRects(startX, endX, startY, endY, CELL_SIZE, rl.Maroon, s.generatedMap.floatingBounceRects[:s.floatingBounceIdx])

			s.square.Draw()
		}
		rl.EndMode2D()
	}
	rl.EndDrawing()
}

func (s *Simulation) Run() {
	s.startTimeSec = rl.GetTime() + START_DELAY_SEC

	for !rl.WindowShouldClose() {
		s.update()
		s.draw()
	}

	rl.CloseAudioDevice()
	rl.UnloadMusicStream(s.music)
	rl.CloseWindow()
}
