package main

import (
	"runtime"
	"time"
	"unsafe"

	"github.com/remogatto/mandala"
	lib "github.com/remogatto/mandala-examples/chipmunk/src/chipmunklib"
	gl "github.com/remogatto/opengles2"
	"github.com/tideland/goas/v2/loop"
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

func newRenderLoopControl() *renderLoopControl {
	return &renderLoopControl{
		make(chan mandala.PauseEvent),
		make(chan bool),
		make(chan initData, 1),
		make(chan [2]float32),
	}
}

// Run runs renderLoop. The loop renders a frame and swaps the buffer
// at each tick received.
func renderLoopFunc(control *renderLoopControl) loop.LoopFunc {
	return func(loop loop.Loop) error {

		var state *lib.GameState

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

				state = lib.NewGameState(window)

				width, height := window.GetSize()
				gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

				ShowAdPopup(activity)

				ticker = time.NewTicker(time.Duration(time.Second / time.Duration(FramesPerSecond)))
				fpsTicker = time.NewTicker(time.Duration(time.Second))

			case tap := <-control.tapEvent:
				state.World.Explosion(tap[0], tap[1])

			// At each tick render a frame and swap buffers.
			case <-ticker.C:
				state.Frames++
				state.Draw()
				state.SwapBuffers()

			case <-fpsTicker.C:
				state.Fps = state.Frames
				state.Frames = 0

			case event := <-control.pause:
				ticker.Stop()
				fpsTicker.Stop()
				state.World.Destroy()
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
