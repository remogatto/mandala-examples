package cubelib

import (
	"github.com/mortdeus/mathgl"
	gl "github.com/remogatto/opengles2"
	"math"
)

type Camera struct {
	X, Y, Z float32
}

type World struct {
	camera           Camera
	objects          []*Cube
	view, projection mathgl.Mat4
	projectionView   mathgl.Mat4
}

func makeFrustum(left, right, bottom, top, nearZ, farZ float32) (mat mathgl.Mat4) {
	deltaX := right - left
	deltaY := top - bottom
	deltaZ := farZ - nearZ

	mat.Fill(0)

	// I column
	mat[0] = 2.0 * nearZ / deltaX

	// II column
	mat[5] = 2.0 * nearZ / deltaY

	// III column
	mat[8] = (right + left) / deltaX
	mat[9] = (top + bottom) / deltaY
	mat[10] = -(nearZ + farZ) / deltaZ
	mat[11] = -1.0

	// IV column
	mat[14] = -2.0 * nearZ * farZ / deltaZ

	return
}

func makePerspective(fov, aspect, nearZ, farZ float32) (mat mathgl.Mat4) {
	frustumH := (float32)(math.Tan(float64(fov/360*mathgl.PI))) * nearZ
	frustumW := frustumH * aspect
	return makeFrustum(-frustumW, frustumW, -frustumH, frustumH, nearZ, farZ)
}

func makeView() (mat mathgl.Mat4) {
	mat.Identity()
	return mat
}

func NewWorld() *World {
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Viewport(0, 0, gl.Sizei(INITIAL_WINDOW_WIDTH), gl.Sizei(INITIAL_WINDOW_HEIGHT))

	return &World{
		objects:    make([]*Cube, 0),
		projection: makePerspective(60.0, 640.0/480.0, 1.0, 20.0),
		view:       makeView(),
	}
}

func (w *World) SetCamera(x, y, z float32) {
	// set the view matrix
	w.view.Translation(-x, -y, -z)
}

func (w *World) Attach(obj *Cube) {
	w.projectionView.Assign(&w.projection)
	w.projectionView.Multiply(&w.view)
	obj.projectionView.Assign(&w.projectionView)
	w.objects = append(w.objects, obj)
}

func (w *World) Draw() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, obj := range w.objects {
		obj.Draw()
	}
}

