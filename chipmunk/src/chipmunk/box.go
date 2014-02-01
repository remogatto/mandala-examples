package main

import (
	"github.com/remogatto/shapes"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

const (
	BoxMass       = 5.0
	BoxElasticity = 0.6

	// BoxSize is in pixel²
	BoxSize = 50 * 50
)

type box struct {
	// Chipumunk stuff
	physicsBody  *chipmunk.Body
	physicsShape *chipmunk.Shape

	// OpenGL stuff
	openglShape *shapes.Box

	world *world
}

func newBox(width, height float32) *box {
	box := new(box)

	// Chipmunk body

	box.physicsShape = chipmunk.NewBox(
		vect.Vect{0, 0},
		vect.Float(width),
		vect.Float(height),
	)

	box.physicsShape.SetElasticity(BoxElasticity)
	box.physicsBody = chipmunk.NewBody(vect.Float(BoxMass), box.physicsShape.Moment(float32(BoxMass)))
	box.physicsBody.AddShape(box.physicsShape)

	// OpenGL shape

	box.openglShape = shapes.NewBox(width, height)

	return box
}

func (box *box) draw() {
	pos := box.physicsBody.Position()
	rot := box.physicsBody.Angle() * chipmunk.DegreeConst
	box.openglShape.Position(float32(pos.X), float32(pos.Y))
	box.openglShape.Rotate(float32(rot))
	box.openglShape.Draw()
}

func (box *box) inViewport() bool {
	pos := box.physicsBody.Position()
	width, _ := box.openglShape.GetSize()
	return float32(pos.X) > -width && float32(pos.X) < (width+float32(box.world.width))
}
