// +build !android

package main

import (
	"flag"
	"git.tideland.biz/goas/loop"
	"github.com/remogatto/gorgasm"
	"strconv"
	"strings"
)

func main() {
	verbose := flag.Bool("verbose", false, "produce verbose output")
	debug := flag.Bool("debug", false, "produce debug output")
	size := flag.String("size", "320x480", "set the size of the window")

	flag.Parse()

	if *verbose {
		gorgasm.Verbose = true
	}

	if *debug {
		gorgasm.Debug = true
	}

	dims := strings.Split(strings.ToLower(*size), "x")
	width, err := strconv.Atoi(dims[0])
	if err != nil {
		panic(err)
	}
	height, err := strconv.Atoi(dims[1])
	if err != nil {
		panic(err)
	}

	// Initialize EGL for xorg.
	gorgasm.XorgInitialize(width, height)

	// Create rendering loop control channels
	renderLoopControl := newRenderLoopControl()

	// Start the rendering loop
	loop.Go(renderLoopFunc(renderLoopControl))

	// Start the event loop
	eventLoop := loop.Go(eventLoopFunc(renderLoopControl))

	// Wait until the event loop exits (i.e. when it receives a
	// DestroyEvent from the Events channel). The event loop is
	// responsible to stop the render loop as well.
	eventLoop.Wait()
}
