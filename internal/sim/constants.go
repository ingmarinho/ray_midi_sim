package sim

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// general
const WINDOW_WIDTH = 720
const WINDOW_HEIGHT = 1280

var WINDOW_CENTER_VECTOR = rl.NewVector2(WINDOW_WIDTH/2, WINDOW_HEIGHT/2)

const FPS = 165
const FRAME_INCREMENT = 1.0 / FPS

// simulation related
const START_DELAY_SEC = 3.0

const SQUARE_SIZE = 50
const SQUARE_SPEED = 400

const BOUNCE_RECT_HEIGHT = 30
const BOUNCE_RECT_WIDTH = 10

// map related
const CELL_SIZE = 10 // has to be a factor of SQUARE_SIZE
const CELL_WAVE_RANGE = 300

const CHANGE_DIR_CHANCE = 0.5

const BACKTRACK_CHANCE = 0.2
const BACKTRACK_AMOUNT = 40
const MAX_RECURSION_DEPTH = 10_000_000

// midi related
const CLOSENESS_THRESHOLD_MS = 1
