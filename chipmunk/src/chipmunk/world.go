package main

import (
	"github.com/remogatto/mathgl"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

const (
	// Y-component for gravity
	Gravity = -900
)

type world struct {
	width, height int
	projMatrix    mathgl.Mat4f
	viewMatrix    mathgl.Mat4f
	space         *chipmunk.Space
	boxes         []*box
	ground        *ground
}

func newWorld(width, height int) *world {
	world := &world{
		width:      width,
		height:     height,
		projMatrix: mathgl.Ortho2D(0, float32(width), 0, float32(height)),
		viewMatrix: mathgl.Ident4f(),
		space:      chipmunk.NewSpace(),
	}
	world.space.Gravity = vect.Vect{0, Gravity}
	return world
}

func (w *world) Projection() mathgl.Mat4f {
	return w.projMatrix
}

func (w *world) View() mathgl.Mat4f {
	return w.viewMatrix
}

func (w *world) addBox(box *box) *box {
	w.space.AddBody(box.physicsBody)
	box.openglShape.AttachToWorld(w)
	w.boxes = append(w.boxes, box)
	return box
}

func (w *world) setGround(ground *ground) *ground {
	w.space.AddBody(ground.physicsBody)
	ground.openglShape.AttachToWorld(w)
	w.ground = ground
	return ground
}
