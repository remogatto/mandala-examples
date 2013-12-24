package main

import (
	"github.com/remogatto/application"
	"github.com/remogatto/egl"
	"github.com/remogatto/egl/platform"
	"github.com/remogatto/gorgasm"
	"github.com/remogatto/gorgasm-examples/cubelib"
	gl "github.com/remogatto/opengles2"
	"image"
	"image/png"
	"runtime"
	"time"
)

const (
	FRAMES_PER_SECOND = 24
)

// renderLoop renders the current scene at a given frame rate.
type renderLoop struct {
	pause, terminate, resume chan int
	ticker                   *time.Ticker
	eglState                 platform.EGLState
}

// newRenderLoop returns a new renderLoop instance. It takes the
// number of frame-per-second as argument.
func newRenderLoop(eglState platform.EGLState, fps int) *renderLoop {
	renderLoop := &renderLoop{
		pause:     make(chan int),
		terminate: make(chan int),
		resume:    make(chan int),
		ticker:    time.NewTicker(time.Duration(1e9 / int(fps))),
		eglState:  eglState,
	}

	return renderLoop
}

// Pause returns the pause channel of the loop.
// If a value is sent to this channel, the loop will be paused.
func (l *renderLoop) Pause() chan int {
	return l.pause
}

// Terminate returns the terminate channel of the loop.
// If a value is sent to this channel, the loop will be terminated.
func (l *renderLoop) Terminate() chan int {
	return l.terminate
}

// Run runs renderLoop. The loop renders a frame and swaps the buffer
// at each tick received.
func (l *renderLoop) Run() {
	// Lock/unlock the loop to the current OS thread. This is
	// necessary because OpenGL functions should be called from
	// the same thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	display := l.eglState.Display
	surface := l.eglState.Surface
	context := l.eglState.Context
	width := l.eglState.SurfaceWidth
	height := l.eglState.SurfaceHeight

	if ok := egl.MakeCurrent(display, surface, surface, context); !ok {
		panic(egl.NewError(egl.GetError()))
	}

	// Create the 3D world
	world := cubelib.NewWorld(width, height)
	world.SetCamera(0.0, 0.0, 5.0)

	cube := cubelib.NewCube()

	img, err := loadImage("res/drawable/marmo.png")
	if err != nil {
		panic(err)
	}
	cube.AttachTexture(img)

	world.Attach(cube)
	angle := float32(0.0)

	for {
		select {

		// Pause the loop.
		case <-l.pause:
			l.ticker.Stop()
			l.pause <- 0

			// Terminate the loop.
		case <-l.terminate:
			l.terminate <- 0

			// Resume the loop.
		case <-l.resume:
			// Do something when the rendering loop is
			// resumed.

			// At each tick render a frame and swap buffers.
		case <-l.ticker.C:
			angle += 0.05
			cube.RotateY(angle)
			world.Draw()
			cubelib.Swap(display, surface)
		}
	}
}

// eventsLoop receives events from the framework and reacts
// accordingly.
type eventsLoop struct {
	pause, terminate chan int
	renderLoop       *renderLoop
}

// newEventsLoop returns a new eventsLoop instance. It takes a
// renderLoop instance as argument.
func newEventsLoop(renderLoop *renderLoop) *eventsLoop {
	eventsLoop := &eventsLoop{
		pause:      make(chan int),
		terminate:  make(chan int),
		renderLoop: renderLoop,
	}
	return eventsLoop
}

// Pause returns the pause channel of the loop.
// If a value is sent to this channel, the loop will be paused.
func (l *eventsLoop) Pause() chan int {
	return l.pause
}

// Terminate returns the terminate channel of the loop.
// If a value is sent to this channel, the loop will be terminated.
func (l *eventsLoop) Terminate() chan int {
	return l.terminate
}

// Run runs eventsLoop listening to events originating from the
// framwork.
func (l *eventsLoop) Run() {
	for {
		select {
		case <-l.pause:
			l.pause <- 0
		case <-l.terminate:
			l.terminate <- 0

			// Receive events from the framework.
		case untypedEvent := <-gorgasm.Events:
			switch event := untypedEvent.(type) {

			// Finger down/up on the screen.
			case gorgasm.ActionUpDownEvent:
				if event.Down {
					application.Logf("Finger is DOWN at coord %d %d", event.X, event.Y)
				} else {
					application.Logf("Finger is now UP")
				}

				// Finger is moving on the screen.
			case gorgasm.ActionMoveEvent:
				application.Logf("Finger is moving at coord %d %d", event.X, event.Y)

			case gorgasm.PauseEvent:
				application.Logf("Application was paused. Stopping rendering ticker.")
				l.renderLoop.pause <- 1

			case gorgasm.ResumeEvent:
				application.Logf("Application was resumed. Reactivating rendering ticker.")
				l.renderLoop.resume <- 1

			}
		}
	}
}

func check() {
	error := gl.GetError()
	if error != 0 {
		application.Logf("An error occurred! Code: 0x%x", error)
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
	gorgasm.Assets <- request
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
