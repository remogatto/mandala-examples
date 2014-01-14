package main

import (
	"image"
	"image/png"
	"runtime"
	"time"

	"git.tideland.biz/goas/loop"
	"github.com/remogatto/mandala"
	"github.com/remogatto/mandala-examples/cube/src/cubelib"
	gl "github.com/remogatto/opengles2"
)

const (
	FRAMES_PER_SECOND = 30
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

type renderState struct {
	window            mandala.Window
	world             *cubelib.World
	cube              *cubelib.Cube
	angle, savedAngle float32
}

func (renderState *renderState) init(window mandala.Window) {
	window.MakeContextCurrent()

	renderState.window = window
	width, height := window.GetSize()

	// Create the 3D world
	renderState.world = cubelib.NewWorld(width, height)
	renderState.world.SetCamera(0.0, 0.0, 5.0)

	renderState.cube = cubelib.NewCube()

	img, err := loadImage("res/drawable/marmo.png")
	if err != nil {
		panic(err)
	}

	renderState.cube.AttachTexture(img)

	renderState.world.Attach(renderState.cube)
	renderState.angle = 0.0
}

func newRenderLoopControl() *renderLoopControl {
	return &renderLoopControl{
		make(chan viewportSize),
		make(chan bool),
		make(chan bool),
		make(chan mandala.Window, 1),
	}
}

// Run runs renderLoop. The loop renders a frame and swaps the buffer
// at each tick received.
func renderLoopFunc(control *renderLoopControl) loop.LoopFunc {
	return func(loop loop.Loop) error {

		// renderState keeps the rendering state variables
		// such as the EGL status
		renderState := new(renderState)

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
				renderState.init(window)
				mandala.Logf("Restarting rendering loop...")
				ticker = time.NewTicker(time.Duration(1e9 / int(FRAMES_PER_SECOND)))

			// At each tick render a frame and swap buffers.
			case <-ticker.C:
				renderState.angle += 0.05
				renderState.cube.Rotate(renderState.angle, [3]float32{0, 1, 0})
				renderState.world.Draw()
				renderState.window.SwapBuffers()

			case viewport := <-control.resizeViewport:
				if renderState.world != nil {
					if viewport.width != renderState.world.Width || viewport.height != renderState.world.Height {
						mandala.Logf("Resize native window W:%v H:%v\n", viewport.width, viewport.height)
						ticker.Stop()
						renderState.world.Resize(viewport.width, viewport.height)
						ticker = time.NewTicker(time.Duration(1e9 / int(FRAMES_PER_SECOND)))
					}
				}

			case <-control.pause:
				renderState.savedAngle = renderState.angle
				mandala.Logf("Save an angle value of %f", renderState.savedAngle)
				ticker.Stop()

			case <-control.resume:
				renderState.angle = renderState.savedAngle
				mandala.Logf("Restore an angle value of %f", renderState.angle)

			case <-loop.ShallStop():
				ticker.Stop()
				return nil
			}
		}
	}
}

// eventLoopFunc listen to events originating from the
// framwork.
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
					width, height := event.Window.GetSize()
					renderLoopControl.resizeViewport <- viewportSize{width, height}

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

func check() {
	error := gl.GetError()
	if error != 0 {
		mandala.Logf("An error occurred! Code: 0x%x", error)
	}
}

func loadImage(filename string) (image.Image, error) {
	// Request an asset to the asset manager. When the app runs on
	// an Android device, the apk will be unpacked and the file
	// will be read from it and copied to a byte buffer.
	request := mandala.LoadAssetRequest{
		filename,
		make(chan mandala.LoadAssetResponse),
	}
	mandala.AssetManager() <- request
	response := <-request.Response

	if response.Error != nil {
		return nil, response.Error
	}

	// Decode the image.
	img, err := png.Decode(response.Buffer)
	if err != nil {
		return nil, err
	}
	return img, nil
}
