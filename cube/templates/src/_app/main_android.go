// +build android

package main

import (
	"fmt"
	"git.tideland.biz/goas/loop"
	"github.com/remogatto/gorgasm"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	gorgasm.Verbose = true
	gorgasm.Debug = true

	// Create rendering loop control channels
	renderLoopControl := newRenderLoopControl()
	// Start the rendering loop
	loop.GoRecoverable(
		renderLoopFunc(renderLoopControl),
		func(rs loop.Recoverings) (loop.Recoverings, error) {
			for _, r := range rs {
				gorgasm.Logf("%s", r.Reason)
				gorgasm.Logf("%s", gorgasm.Stacktrace())
			}
			return rs, fmt.Errorf("Unrecoverable loop\n")
		},
	)
	// Start the event loop
	loop.GoRecoverable(
		eventLoopFunc(renderLoopControl),
		func(rs loop.Recoverings) (loop.Recoverings, error) {
			for _, r := range rs {
				gorgasm.Logf("%s", r.Reason)
				gorgasm.Logf("%s", gorgasm.Stacktrace())
			}
			return rs, fmt.Errorf("Unrecoverable loop\n")
		},
	)

}
