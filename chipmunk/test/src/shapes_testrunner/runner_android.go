// +build android

package main

import (
	"runtime"

	"testlib"
	"github.com/remogatto/mandala"
	mandalatest "github.com/remogatto/mandala/test/src/testlib"
	"github.com/remogatto/prettytest"
)

const (
	outputPath = "/data/data/net.mandala.test.shapes_testrunner/files/"
)

type T struct{}

func (t T) Fail() {}

var t T

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	mandala.Verbose = true

	go prettytest.RunWithFormatter(
		t,
		new(mandalatest.TDDFormatter),
		testlib.NewTestSuite(outputPath),
	)
}
