package main

import (
	"image/color"

	"github.com/remogatto/shapes"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

const (
	GroundRadius = 1.0
)

type ground struct {
	physicsShape *chipmunk.Shape
	physicsBody  *chipmunk.Body
	openglShape  *shapes.Segment
}

func newGround(x1, y1, x2, y2 float32) *ground {
	ground := new(ground)

	// Chipmunk body

	ground.physicsBody = chipmunk.NewBodyStatic()
	ground.physicsShape = chipmunk.NewSegment(
		vect.Vect{vect.Float(x1), vect.Float(y1)},
		vect.Vect{vect.Float(x2), vect.Float(y2)},
		GroundRadius,
	)

	ground.physicsBody.AddShape(ground.physicsShape)

	// OpenGL shape

	ground.openglShape = shapes.NewSegment(x1, y1, x2, y2)
	ground.openglShape.Color(color.White)

	return ground
}

func (ground *ground) draw() {
	ground.openglShape.Draw()
}
