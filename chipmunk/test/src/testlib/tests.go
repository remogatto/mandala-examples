package testlib

import (
	"fmt"
	"math/rand"

	"github.com/remogatto/imagetest"
	lib "github.com/remogatto/mandala-examples/chipmunk/src/chipmunklib"
	"github.com/remogatto/mandala/test/src/testlib"
)

const distanceThreshold = 0.002

func distanceError(distance float64, filename string) string {
	return fmt.Sprintf("Image differs by distance %f, result saved in %s", distance, filename)
}

func (t *TestSuite) TestRenderWorld() {
	filename := "expected_world.png"
	t.rlControl.drawFunc <- func() {

		// initialize the default source to a deterministic
		// state
		rand.Seed(1234)

		state := lib.NewGameState(t.renderState.window)
		state.Draw()
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
