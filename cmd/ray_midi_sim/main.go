package main

import (
	"ray_midi_sim/internal/sim"
)

// IDEAS
// - warped grid effect
// - grid colors that move and change on bounce
// - start animation mini grid turns into square
// - multiple squares in the same map, generate multi square map (with max distance between squares)
// - multiple maps at the same time
// - connect bounce rects with outer grid based on distance (possible collisions could be checked using path polygons)
// - particles

func main() {
	midPath := `C:\Users\ingma\Desktop\RhythmVisualizer\18.03.25\__Maretu2.mid` // `C:\Users\ingma\Desktop\mids\Transcribed_ Calix Huang - Carry You Home.mid`
	wavFilePath := `C:\Users\ingma\Desktop\RhythmVisualizer\18.03.25\__Maretu2.wav`

	sim := sim.New(midPath, wavFilePath)

	err := sim.Init()
	if err != nil {
		panic(err)
	}

	sim.Run()
}
