package testlib

import (
	"fmt"

	"github.com/remogatto/imagetest"
	"github.com/remogatto/mandala/test/src/testlib"
)

const (
	distanceThreshold = 0.002
	texFilename       = "gopher.png"
	texDistThreshold  = 0.004
)

func distanceError(distance float64, filename string) string {
	return fmt.Sprintf("Image differs by distance %f, result saved in %s", distance, filename)
}

func (t *TestSuite) TestRenderWorld() {
	filename := "expected_world.png"
	t.rlControl.drawFunc <- func() {
		// w, h := t.renderState.window.GetSize()
		// world := newWorld(w, h)
		// // Create a box
		// box := shapes.NewBox(t.renderState.boxProgram, 100, 100)
		// box.AttachToWorld(world)
		// gl.Clear(gl.COLOR_BUFFER_BIT)
		// box.MoveTo(float32(w/2), 0)
		// box.Draw()
		state := engine.NewGameState()
		state.World.Load("world.svg")
		state.World.Draw()
		t.testDraw <- testlib.Screenshot(t.renderState.window)
		t.renderState.window.SwapBuffers()
	}
	distance, exp, act, err := testlib.TestImage(filename, <-t.testDraw, imagetest.Center)
	if err != nil {
		panic(err)
	}
	t.True(distance < distanceThreshold, distanceError(distance, filename))
	if t.Failed() {
		saveExpAct(t.outputPath, "failed_"+filename, exp, act)
	}
}
