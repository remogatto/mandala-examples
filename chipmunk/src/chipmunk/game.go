package main

import (
	"math/rand"
	"runtime"
	"time"
	"unsafe"

	"github.com/remogatto/mandala"
	gl "github.com/remogatto/opengles2"
	"github.com/tideland/goas/v2/loop"
	"github.com/vova616/chipmunk/vect"
)

const (
	FramesPerSecond = 30
	NumOfBoxes      = 50
)

type initData struct {
	window   mandala.Window
	activity unsafe.Pointer
}

type renderLoopControl struct {
	pause    chan mandala.PauseEvent
	resume   chan bool
	init     chan initData
	tapEvent chan [2]float32
}

type gameState struct {
	window      mandala.Window
	world       *world
	fps, frames int
}

func newGameState(window mandala.Window) *gameState {
	s := new(gameState)
	s.window = window

	s.window.MakeContextCurrent()

	w, h := window.GetSize()

	s.world = newWorld(w, h)

	// Create the building reading it from a string
	rand.Seed(int64(time.Now().Nanosecond()))

	// Uncomment the following lines to generate the world
	// starting from a string (defined in world.go)

	// s.world.createFromString(pyramid)
	// s.world.setGround(newGround(0, float32(10), float32(w), float32(10)))

	s.world.createFromSvg("raw/world.svg")

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	return s
}

func newRenderLoopControl() *renderLoopControl {
	return &renderLoopControl{
		make(chan mandala.PauseEvent),
		make(chan bool),
		make(chan initData, 1),
		make(chan [2]float32),
	}
}

func (s *gameState) printFPS(x, y float32) {
	text, err := s.world.font.Printf("Frames per second %d", s.fps)
	if err != nil {
		panic(err)
	}
	text.AttachToWorld(s.world)
	text.MoveTo(x, y)
	text.Draw()
}

func (s *gameState) draw() {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	s.world.space.Step(vect.Float(1 / float32(FramesPerSecond)))

	for i := 0; i < len(s.world.boxes); i++ {
		box := s.world.boxes[i]
		if box.inViewport() {
			box.draw()
		} else {
			s.world.removeBox(box, i)
			i--
		}
	}

	s.world.ground.draw()

	s.printFPS(float32(s.world.width/2), float32(s.world.height)-25)
}

// Run runs renderLoop. The loop renders a frame and swaps the buffer
// at each tick received.
func renderLoopFunc(control *renderLoopControl) loop.LoopFunc {
	return func(loop loop.Loop) error {

		var state *gameState

		// Lock/unlock the loop to the current OS thread. This is
		// necessary because OpenGL functions should be called from
		// the same thread.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// Create an instance of ticker and immediately stop
		// it because we don't want to swap buffers before
		// initializing a rendering state.
		ticker := time.NewTicker(time.Duration(1e9 / int(FramesPerSecond)))
		ticker.Stop()

		fpsTicker := time.NewTicker(time.Duration(time.Second))
		fpsTicker.Stop()

		for {
			select {
			case init := <-control.init:
				window := init.window
				activity := init.activity

				ticker.Stop()

				state = newGameState(window)

				width, height := window.GetSize()
				gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

				ShowAdPopup(activity)

				ticker = time.NewTicker(time.Duration(time.Second / time.Duration(FramesPerSecond)))
				fpsTicker = time.NewTicker(time.Duration(time.Second))

			case tap := <-control.tapEvent:
				state.world.explosion(tap[0], tap[1])

			// At each tick render a frame and swap buffers.
			case <-ticker.C:
				state.frames++
				state.draw()
				state.window.SwapBuffers()

			case <-fpsTicker.C:
				state.fps = state.frames
				state.frames = 0

			case event := <-control.pause:
				ticker.Stop()
				fpsTicker.Stop()
				state.world.destroy()
				event.Paused <- true

			case <-control.resume:

			case <-loop.ShallStop():
				ticker.Stop()
				return nil
			}
		}
	}
}

// eventLoopFunc listen to events originating from the
// framework.
func eventLoopFunc(renderLoopControl *renderLoopControl) loop.LoopFunc {
	return func(loop loop.Loop) error {

		for {
			select {

			case untypedEvent := <-mandala.Events():
				switch event := untypedEvent.(type) {

				// Receive a native window
				// from the framework and send
				// it to the render loop in
				// order to begin the
				// rendering process.
				case mandala.NativeWindowCreatedEvent:
					renderLoopControl.init <- initData{event.Window, event.Activity}

				// Finger down/up on the screen.
				case mandala.ActionUpDownEvent:
					if event.Down {
						renderLoopControl.tapEvent <- [2]float32{event.X, event.Y}
					}

					// Finger is moving on the screen.
				case mandala.ActionMoveEvent:
					mandala.Logf("Finger is moving at coord %f %f", event.X, event.Y)

				case mandala.DestroyEvent:
					mandala.Logf("Quitting from application now...\n")
					return nil

				case mandala.NativeWindowRedrawNeededEvent:

				case mandala.PauseEvent:
					mandala.Logf("Application was paused. Stopping rendering ticker.")
					renderLoopControl.pause <- event

				case mandala.ResumeEvent:
					mandala.Logf("Application was resumed. Reactivating rendering ticker.")
					renderLoopControl.resume <- true

				}
			}
		}
	}
}
