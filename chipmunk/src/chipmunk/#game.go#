package main

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"time"

	"git.tideland.biz/goas/loop"
	"github.com/remogatto/mandala"
	"github.com/remogatto/mathgl"
	gl "github.com/remogatto/opengles2"
	"github.com/remogatto/shapes"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

const (
	FRAMES_PER_SECOND = 24
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

type world struct {
	width, height int
	projMatrix    mathgl.Mat4f
	viewMatrix    mathgl.Mat4f
}

type gameState struct {
	window   mandala.Window
	space    *chipmunk.Space
	ground   *chipmunk.Shape
	box      *chipmunk.Shape
	glBox    *shapes.Box
	glGround *shapes.Line
	world    *world
}

func newWorld(width, height int) *world {
	return &world{
		width:      width,
		height:     height,
		projMatrix: mathgl.Ortho2D(0, float32(width), 0, float32(height)),
		viewMatrix: mathgl.Ident4f(),
	}
}

func (w *world) Projection() mathgl.Mat4f {
	return w.projMatrix
}

func (w *world) View() mathgl.Mat4f {
	return w.viewMatrix
}

func newGameState(window mandala.Window) *gameState {
	s := new(gameState)
	s.window = window
	w, h := window.GetSize()

	s.window.MakeContextCurrent()

	// Create OpenGL world

	s.world = newWorld(w, h)

	// Create chipmunk space

	s.space = chipmunk.NewSpace()
	s.space.Gravity = vect.Vect{0, -2000}

	// Add the ground
	staticBody := chipmunk.NewBodyStatic()
	s.ground = chipmunk.NewSegment(vect.Vect{0, vect.Float(h) / 2}, vect.Vect{320, vect.Float(h) / 2}, 1)
	staticBody.AddShape(s.ground)
	s.space.AddBody(staticBody)

	// Add a box
	boxMass := 1.0
	s.box = chipmunk.NewBox(
		vect.Vect{0, 0},
		vect.Float(50),
		vect.Float(50),
	)

	s.box.SetElasticity(0.8)
	body := chipmunk.NewBody(vect.Float(boxMass), s.box.Moment(float32(boxMass)))
	body.SetPosition(vect.Vect{vect.Float(w / 2), vect.Float(h)})
	body.SetAngle(vect.Float(rand.Float32() * 2 * math.Pi))

	body.AddShape(s.box)
	s.space.AddBody(body)

	// OpenGL shapes

	s.glBox = shapes.NewBox(50, 50)
	s.glBox.AttachToWorld(s.world)

	s.glGround = shapes.NewLine(0, float32(h/2), float32(w), float32(h/2))
	s.glGround.AttachToWorld(s.world)
	s.glGround.Color(color.White)

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

	s.space.Step(vect.Float(1 / float32(FRAMES_PER_SECOND)))

	pos := s.box.Body.Position()
	rot := s.box.Body.Angle() * chipmunk.DegreeConst
	s.glBox.Position(float32(pos.X), float32(pos.Y))
	s.glBox.Rotate(float32(rot))
	s.glBox.Draw()
	s.glGround.Draw()
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
		ticker := time.NewTicker(time.Duration(1e9 / int(FRAMES_PER_SECOND)))
		ticker.Stop()

		for {
			select {
			case window := <-control.window:
				ticker.Stop()

				state = newGameState(window)

				width, height := window.GetSize()
				gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

				mandala.Logf("Restarting rendering loop...")
				ticker = time.NewTicker(time.Duration(1e9 / int(FRAMES_PER_SECOND)))

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

			// Receive an EGL state from the
			// framework and notify the render
			// loop about that.
			// case eglState := <-mandala.Init:
			// 	mandala.Logf("EGL surface initialized W:%d H:%d", eglState.SurfaceWidth, eglState.SurfaceHeight)
			// 	renderLoopControl.eglState <- eglState

			// Receive events from the framework.
			//
			// When the application starts the
			// typical events chain is:
			//
			// * onCreate
			// * onResume
			// * onInputQueueCreated
			// * onNativeWindowCreated
			// * onNativeWindowResized
			// * onWindowFocusChanged
			// * onNativeRedrawNeeded
			//
			// Pausing (i.e. clicking on the back
			// button) the application produces
			// following events chain:
			//
			// * onPause
			// * onWindowDestroy
			// * onWindowFocusChanged
			// * onInputQueueDestroy
			// * onDestroy

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
