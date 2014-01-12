package cubelib

import (
	"github.com/remogatto/mathgl"
	gl "github.com/remogatto/opengles2"
)

type Camera struct {
	X, Y, Z float32
}

type World struct {
	// Size of the rendering viewport
	Width, Height int

	camera           Camera
	objects          []*Cube
	view, projection mathgl.Mat4f
	projectionView   mathgl.Mat4f
}

func NewWorld(width, height int) *World {
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

	return &World{
		Width:      width,
		Height:     height,
		objects:    make([]*Cube, 0),
		projection: mathgl.Perspective(60, float32(width)/float32(height), 1, 20),
		view:       mathgl.Ident4f(),
	}
}

func (w *World) Resize(width, height int) {
	w.Width, w.Height = width, height
	gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))
	w.projection = mathgl.Perspective(60, float32(width)/float32(height), 1, 20)
	w.projectionView = w.projection
	w.projectionView.Mul4(w.view)
	for _, obj := range w.objects {
		obj.projectionView = w.projectionView
	}
}

func (w *World) SetCamera(x, y, z float32) {
	// set the view matrix
	w.view = mathgl.Translate3D(-x, -y, -z)
}

func (w *World) Attach(obj *Cube) {
	w.projectionView = w.projection
	w.projectionView = w.projectionView.Mul4(w.view)
	obj.projectionView = w.projectionView
	w.objects = append(w.objects, obj)
}

func (w *World) Draw() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, obj := range w.objects {
		obj.Draw()
	}
}
