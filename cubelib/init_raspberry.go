// +build raspberry

package cubelib

import (
	"github.com/remogatto/egl"
	"github.com/remogatto/egl/platform/raspberry"
)

const (
	INITIAL_WINDOW_WIDTH  = 1920
	INITIAL_WINDOW_HEIGHT = 1080
)

func Initialize() {
	egl.BCMHostInit()
	raspberry.Initialize(raspberry.DefaultConfigAttributes, raspberry.DefaultContextAttributes)
}
