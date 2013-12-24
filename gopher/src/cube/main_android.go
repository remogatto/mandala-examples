// +build android

package main

import (
	"github.com/remogatto/application"
	"github.com/remogatto/gorgasm"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	application.Verbose = true
	application.Debug = true

	// The main() function can't be blocking on Android so we have
	// to launch the application control loop on a separate
	// goroutine.
	go func() {
		for {
			select {
			case eglState := <-gorgasm.Init:
				renderLoop := newRenderLoop(eglState, FRAMES_PER_SECOND)
				eventsLoop := newEventsLoop(renderLoop)
				application.Register("renderLoop", renderLoop)
				application.Register("eventsLoop", eventsLoop)
				application.Start("renderLoop")
				application.Start("eventsLoop")
			case <-application.Exit:
				return
			case err := <-application.Errors:
				application.Logf(err.(application.Error).Error())
			}
		}
	}()
}
