package main

import (
	"git.tideland.biz/goas/loop"
	"github.com/remogatto/egl"
	"github.com/remogatto/egl/platform"
	"github.com/remogatto/gorgasm"
	"github.com/remogatto/gorgasm-examples/cube/src/cubelib"
	gl "github.com/remogatto/opengles2"
	"image"
	"image/png"
	"runtime"
	"time"
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
	eglState       chan *platform.EGLState
}

type renderState struct {
	eglState          *platform.EGLState
	world             *cubelib.World
	cube              *cubelib.Cube
	angle, savedAngle float32
}

func (renderState *renderState) init(eglState *platform.EGLState) {
	renderState.eglState = eglState

	display := eglState.Display
	surface := eglState.Surface
	context := eglState.Context
	width := eglState.SurfaceWidth
	height := eglState.SurfaceHeight

	if ok := egl.MakeCurrent(display, surface, surface, context); !ok {
		panic(egl.NewError(egl.GetError()))
	}

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
		make(chan *platform.EGLState),
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
			case eglState := <-control.eglState:
				ticker.Stop()
				renderState.init(eglState)
				gorgasm.Logf("Restarting rendering loop...")
				ticker = time.NewTicker(time.Duration(1e9 / int(FRAMES_PER_SECOND)))

			// At each tick render a frame and swap buffers.
			case <-ticker.C:
				renderState.angle += 0.05
				renderState.cube.Rotate(renderState.angle, [3]float32{0, 1, 0})
				renderState.world.Draw()
				egl.SwapBuffers(renderState.eglState.Display, renderState.eglState.Surface)

			case viewport := <-control.resizeViewport:
				if renderState.world != nil {
					if viewport.width != renderState.world.Width || viewport.height != renderState.world.Height {
						gorgasm.Logf("Resize native window W:%v H:%v\n", viewport.width, viewport.height)
						ticker.Stop()
						renderState.world.Resize(viewport.width, viewport.height)
						ticker = time.NewTicker(time.Duration(1e9 / int(FRAMES_PER_SECOND)))
					}
				}

			case <-control.pause:
				renderState.savedAngle = renderState.angle
				gorgasm.Logf("Save an angle value of %f", renderState.savedAngle)
				ticker.Stop()

			case <-control.resume:
				renderState.angle = renderState.savedAngle
				gorgasm.Logf("Restore an angle value of %f", renderState.angle)

			case <-loop.ShallStop():
				ticker.Stop()
				return nil
			}
		}
	}
}

// eventLoopFunc is listening for events originating from the
// framwork.
func eventLoopFunc(renderLoopControl *renderLoopControl) loop.LoopFunc {
	return func(loop loop.Loop) error {

		for {
			select {

			// Receive an EGL state from the
			// framework and notify the render
			// loop about that.
			case eglState := <-gorgasm.Init:
				gorgasm.Logf("EGL surface initialized W:%d H:%d", eglState.SurfaceWidth, eglState.SurfaceHeight)
				renderLoopControl.eglState <- eglState

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

			case untypedEvent := <-gorgasm.Events:
				switch event := untypedEvent.(type) {

				// Finger down/up on the screen.
				case gorgasm.ActionUpDownEvent:
					if event.Down {
						gorgasm.Logf("Finger is DOWN at %f %f", event.X, event.Y)
					} else {
						gorgasm.Logf("Finger is now UP")
					}

					// Finger is moving on the screen.
				case gorgasm.ActionMoveEvent:
					gorgasm.Logf("Finger is moving at coord %f %f", event.X, event.Y)

				case gorgasm.DestroyEvent:
					gorgasm.Logf("Stop rendering...\n")
					gorgasm.Logf("Quitting from application...\n")
					return nil

				case gorgasm.NativeWindowRedrawNeededEvent:
					width, height := event.Window.Width, event.Window.Height
					renderLoopControl.resizeViewport <- viewportSize{width, height}

				case gorgasm.PauseEvent:
					gorgasm.Logf("Application was paused. Stopping rendering ticker.")
					renderLoopControl.pause <- true

				case gorgasm.ResumeEvent:
					gorgasm.Logf("Application was resumed. Reactivating rendering ticker.")
					renderLoopControl.resume <- true

				}
			}
		}
	}
}

func check() {
	error := gl.GetError()
	if error != 0 {
		gorgasm.Logf("An error occurred! Code: 0x%x", error)
	}
}

func loadImage(filename string) (image.Image, error) {
	// Request an asset to the asset manager. When the app runs on
	// an Android device, the apk will be unpacked and the file
	// will be read from it and copied to a byte buffer.
	request := gorgasm.LoadAssetRequest{
		filename,
		make(chan gorgasm.LoadAssetResponse),
	}
	gorgasm.AssetManager <- request
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
