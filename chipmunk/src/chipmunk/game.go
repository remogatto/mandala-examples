package main

import (
	"math"
	"math/rand"
	"runtime"
	"time"

	"git.tideland.biz/goas/loop"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/remogatto/mandala"
	gl "github.com/remogatto/opengles2"
	"github.com/vova616/chipmunk/vect"
)

const (
	FramesPerSecond = 30
	NumOfBoxes      = 50
)

type viewportSize struct {
	width, height int
}

type renderLoopControl struct {
	resizeViewport chan viewportSize
	pause          chan bool
	resume         chan bool
	window         chan mandala.Window
}

type gameState struct {
	window mandala.Window
	world  *world
}

func newGameState(window mandala.Window) *gameState {
	s := new(gameState)
	s.window = window
	w, h := window.GetSize()

	s.window.MakeContextCurrent()

	s.world = newWorld(w, h)

	// Set the ground

	s.world.setGround(newGround(0, float32(h/3), float32(w), float32(h/3)))

	// Add boxes
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := 0; i < NumOfBoxes; i++ {
		box := s.world.addBox(newBox(40, 40))
		// Initial position and angle of the box
		box.physicsBody.SetPosition(vect.Vect{vect.Float(rand.Float32() * float32(w)), vect.Float(h)})
		box.physicsBody.SetAngle(vect.Float(rand.Float32() * 2 * math.Pi))
		box.openglShape.Color(colorful.HappyColor())
	}

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	return s
}

func newRenderLoopControl() *renderLoopControl {
	return &renderLoopControl{
		make(chan viewportSize),
		make(chan bool),
		make(chan bool),
		make(chan mandala.Window, 1),
	}
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

		for {
			select {
			case window := <-control.window:
				ticker.Stop()

				state = newGameState(window)

				width, height := window.GetSize()
				gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

				ticker = time.NewTicker(time.Duration(time.Second / time.Duration(FramesPerSecond)))

			// At each tick render a frame and swap buffers.
			case <-ticker.C:
				state.draw()
				state.window.SwapBuffers()

			case <-control.pause:
				ticker.Stop()

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
					renderLoopControl.window <- event.Window

				// Finger down/up on the screen.
				case mandala.ActionUpDownEvent:
					if event.Down {
						mandala.Logf("Finger is DOWN at %f %f", event.X, event.Y)
					} else {
						mandala.Logf("Finger is now UP")
					}

					// Finger is moving on the screen.
				case mandala.ActionMoveEvent:
					mandala.Logf("Finger is moving at coord %f %f", event.X, event.Y)

				case mandala.DestroyEvent:
					mandala.Logf("Stop rendering...\n")
					mandala.Logf("Quitting from application...\n")
					return nil

				case mandala.NativeWindowRedrawNeededEvent:

				case mandala.PauseEvent:
					mandala.Logf("Application was paused. Stopping rendering ticker.")
					renderLoopControl.pause <- true

				case mandala.ResumeEvent:
					mandala.Logf("Application was resumed. Reactivating rendering ticker.")
					renderLoopControl.resume <- true

				}
			}
		}
	}
}
