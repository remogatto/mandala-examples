package testlib

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/remogatto/mandala"
	"github.com/remogatto/mathgl"
	gl "github.com/remogatto/opengles2"
	"github.com/remogatto/prettytest"
	"github.com/remogatto/shaders"
	"github.com/remogatto/shapes"
	"github.com/tideland/goas/v2/loop"
)

const (
	// We don't need high framerate for testing
	FramesPerSecond = 15

	expectedImgPath = "drawable"
)

type world struct {
	width, height int
	projMatrix    mathgl.Mat4f
	viewMatrix    mathgl.Mat4f
}

type TestSuite struct {
	prettytest.Suite

	rlControl *renderLoopControl
	timeout   <-chan time.Time

	testDraw chan image.Image

	renderState *renderState
	outputPath  string
}

type renderLoopControl struct {
	window   chan mandala.Window
	drawFunc chan func()
}

type renderState struct {
	window                     mandala.Window
	boxProgram, segmentProgram shaders.Program
}

func (renderState *renderState) init(window mandala.Window) {
	window.MakeContextCurrent()

	renderState.window = window
	width, height := window.GetSize()

	// Set the viewport
	gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	renderState.boxProgram = shaders.NewProgram(shapes.DefaultBoxFS, shapes.DefaultBoxVS)
	renderState.segmentProgram = shaders.NewProgram(shapes.DefaultSegmentFS, shapes.DefaultSegmentVS)
}

func newRenderLoopControl() *renderLoopControl {
	return &renderLoopControl{
		drawFunc: make(chan func()),
		window:   make(chan mandala.Window),
	}
}

// Timeout timeouts the tests after the given duration.
func (t *TestSuite) Timeout(timeout time.Duration) {
	t.timeout = time.After(timeout)
}

// Run runs renderLoop. The loop renders a frame and swaps the buffer
// at each tick received.
func (t *TestSuite) renderLoopFunc(control *renderLoopControl) loop.LoopFunc {
	return func(loop loop.Loop) error {

		// renderState stores rendering state variables such
		// as the EGL state
		t.renderState = new(renderState)

		// Lock/unlock the loop to the current OS thread. This is
		// necessary because OpenGL functions should be called from
		// the same thread.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		window := <-control.window
		t.renderState.init(window)

		for {
			select {
			case drawFunc := <-control.drawFunc:
				drawFunc()
			}
		}
	}
}

// eventLoopFunc is listening for events originating from the
// framwork.
func (t *TestSuite) eventLoopFunc(renderLoopControl *renderLoopControl) loop.LoopFunc {
	return func(loop loop.Loop) error {

		for {
			select {

			// Receive events from the framework.
			case untypedEvent := <-mandala.Events():

				switch event := untypedEvent.(type) {

				case mandala.CreateEvent:

				case mandala.StartEvent:

				case mandala.NativeWindowCreatedEvent:
					renderLoopControl.window <- event.Window

				case mandala.ActionUpDownEvent:

				case mandala.ActionMoveEvent:

				case mandala.NativeWindowDestroyedEvent:

				case mandala.DestroyEvent:

				case mandala.NativeWindowRedrawNeededEvent:

				case mandala.PauseEvent:

				case mandala.ResumeEvent:

				}
			}
		}
	}
}

func (t *TestSuite) timeoutLoopFunc() loop.LoopFunc {
	return func(loop loop.Loop) error {
		time := <-t.timeout
		err := fmt.Errorf("Tests timed out after %v", time)
		mandala.Logf("%s %s", err.Error(), mandala.Stacktrace())
		t.Error(err)
		return nil
	}
}

func (t *TestSuite) BeforeAll() {
	// Create rendering loop control channels
	t.rlControl = newRenderLoopControl()
	// Start the rendering loop
	loop.GoRecoverable(
		t.renderLoopFunc(t.rlControl),
		func(rs loop.Recoverings) (loop.Recoverings, error) {
			for _, r := range rs {
				mandala.Logf("%s", r.Reason)
				mandala.Logf("%s", mandala.Stacktrace())
			}
			return rs, fmt.Errorf("Unrecoverable loop\n")
		},
	)
	// Start the event loop
	loop.GoRecoverable(
		t.eventLoopFunc(t.rlControl),
		func(rs loop.Recoverings) (loop.Recoverings, error) {
			for _, r := range rs {
				mandala.Logf("%s", r.Reason)
				mandala.Logf("%s", mandala.Stacktrace())
			}
			return rs, fmt.Errorf("Unrecoverable loop\n")
		},
	)

	if t.timeout != nil {
		// Start the timeout loop
		loop.GoRecoverable(
			t.timeoutLoopFunc(),
			func(rs loop.Recoverings) (loop.Recoverings, error) {
				for _, r := range rs {
					mandala.Logf("%s", r.Reason)
					mandala.Logf("%s", mandala.Stacktrace())
				}
				return rs, fmt.Errorf("Unrecoverable loop\n")
			},
		)
	}

}

func newWorld(width, height int) *world {
	return &world{
		width:      width,
		height:     height,
		projMatrix: mathgl.Ortho2D(0, float32(width), -float32(height/2), float32(height/2)),
		viewMatrix: mathgl.Ident4f(),
	}
}

func (w *world) Projection() mathgl.Mat4f {
	return w.projMatrix
}

func (w *world) View() mathgl.Mat4f {
	return w.viewMatrix
}

func (w *world) addImageAsTexture(filename string) uint32 {
	var texBuffer uint32
	texImg, err := loadImageResource(filename)
	if err != nil {
		panic(err)
	}
	b := texImg.Bounds()
	rgbaImage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgbaImage, rgbaImage.Bounds(), texImg, b.Min, draw.Src)

	width, height := gl.Sizei(b.Dx()), gl.Sizei(b.Dy())
	gl.GenTextures(1, &texBuffer)
	gl.BindTexture(gl.TEXTURE_2D, texBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Void(&rgbaImage.Pix[0]))

	return texBuffer
}

// loadImageResource loads an image with the given filename from the
// resource folder.
func loadImageResource(filename string) (image.Image, error) {
	request := mandala.LoadResourceRequest{
		Filename: filepath.Join(expectedImgPath, filename),
		Response: make(chan mandala.LoadResourceResponse),
	}

	mandala.ResourceManager() <- request
	response := <-request.Response

	buffer := response.Buffer
	if response.Error != nil {
		return nil, response.Error
	}

	img, err := png.Decode(bytes.NewReader(buffer))

	if err != nil {
		return nil, err
	}

	return img, nil
}

// Create an image containing both expected and actual images, side by
// side.
func saveExpAct(outputPath string, filename string, exp image.Image, act image.Image) {

	// Build the destination rectangle
	expRect := exp.Bounds()
	actRect := act.Bounds()
	unionRect := expRect.Union(actRect)
	dstRect := image.Rectangle{
		image.ZP,
		image.Point{unionRect.Max.X * 3, unionRect.Max.Y},
	}

	// Create the empty destination image
	dstImage := image.NewRGBA(dstRect)

	// Copy the expected image
	dp := image.Point{
		(unionRect.Max.X-unionRect.Min.X)/2 - (expRect.Max.X-expRect.Min.X)/2,
		(unionRect.Max.Y-unionRect.Min.Y)/2 - (expRect.Max.Y-expRect.Min.Y)/2,
	}
	r := image.Rectangle{dp, dp.Add(expRect.Size())}
	draw.Draw(dstImage, r, exp, image.ZP, draw.Src)

	// Copy the actual image
	dp = image.Point{
		(unionRect.Max.X-unionRect.Min.X)/2 - (actRect.Max.X-expRect.Min.X)/2,
		(unionRect.Max.Y-unionRect.Min.Y)/2 - (actRect.Max.Y-expRect.Min.Y)/2,
	}
	dp = dp.Add(image.Point{unionRect.Max.X, 0})
	r = image.Rectangle{dp, dp.Add(actRect.Size())}
	draw.Draw(dstImage, r, act, image.ZP, draw.Src)

	// Re-copy the actual image
	dp = dp.Add(image.Point{unionRect.Max.X, 0})
	r = image.Rectangle{dp, dp.Add(actRect.Size())}
	draw.Draw(dstImage, r, act, image.ZP, draw.Src)

	// Composite expected over actual
	dp = image.Point{dp.X, unionRect.Min.Y}
	dstRect = image.Rectangle{dp, dp.Add(expRect.Size())}
	draw.DrawMask(dstImage, dstRect, exp, image.ZP, &image.Uniform{color.RGBA{A: 64}}, image.ZP, draw.Over)

	_, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		// Create the output dir
		err := os.Mkdir(outputPath, 0777)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	// Save the output file
	file, err := os.Create(filepath.Join(outputPath, filename))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = png.Encode(file, dstImage)
	if err != nil {
		panic(err)
	}
}

func NewTestSuite(outputPath string) *TestSuite {
	return &TestSuite{
		rlControl:  newRenderLoopControl(),
		testDraw:   make(chan image.Image),
		outputPath: outputPath,
	}
}
