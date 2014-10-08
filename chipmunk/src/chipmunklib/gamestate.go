package chipmunklib

import (
	"github.com/remogatto/mandala"
	gl "github.com/remogatto/opengles2"
	"github.com/vova616/chipmunk/vect"
)

const DefaultFps = 30

type GameState struct {
	window      mandala.Window
	World       *World
	Fps, Frames int
}

// NewGameState creates a new game state. It needs a window onto which
// render the scene.
func NewGameState(window mandala.Window) *GameState {
	s := new(GameState)
	s.window = window

	s.window.MakeContextCurrent()

	w, h := window.GetSize()

	s.World = NewWorld(w, h)

	s.Fps = DefaultFps

	// Uncomment the following lines to generate the world
	// starting from a string (defined in world.go)

	// s.world.createFromString(pyramid)
	// s.world.setGround(newGround(0, float32(10), float32(w), float32(10)))

	s.World.CreateFromSvg("raw/world.svg")

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	return s
}

func (s *GameState) printFPS(x, y float32) {
	text, err := s.World.font.Printf("Frames per second %d", s.Fps)
	if err != nil {
		panic(err)
	}
	text.AttachToWorld(s.World)
	text.MoveTo(x, y)
	text.Draw()
}

func (s *GameState) Draw() {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	s.World.space.Step(vect.Float(1 / float32(s.Fps)))

	for i := 0; i < len(s.World.boxes); i++ {
		box := s.World.boxes[i]
		if box.inViewport() {
			box.draw()
		} else {
			s.World.removeBox(box, i)
			i--
		}
	}

	s.World.ground.draw()

	s.printFPS(float32(s.World.width/2), float32(s.World.height)-25)
}

func (s *GameState) SwapBuffers() {
	s.window.SwapBuffers()
}
