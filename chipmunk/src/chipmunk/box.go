package main

import (
	"github.com/remogatto/mandala"
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

type callbacks struct{}

type impactUserData struct {
	player       *mandala.AudioPlayer
	impactBuffer []byte
}

type box struct {
	// Chipumunk stuff
	physicsBody  *chipmunk.Body
	physicsShape *chipmunk.Shape

	// OpenGL stuff
	openglShape *shapes.Box

	// Sound player
	player *mandala.AudioPlayer

	world *world
}

func (c callbacks) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	a := arbiter.BodyA
	b := arbiter.BodyB
	impact, ok := a.UserData.(impactUserData)
	if ok {
		impact.player.Play(impact.impactBuffer, nil)
	}
	impact = b.UserData.(impactUserData)
	if ok {
		impact.player.Play(impact.impactBuffer, nil)
	}
	return true
}

func (c callbacks) CollisionPreSolve(arbiter *chipmunk.Arbiter) bool {
	return true
}

func (c callbacks) CollisionPostSolve(arbiter *chipmunk.Arbiter) {}

func (c callbacks) CollisionExit(arbiter *chipmunk.Arbiter) {}

func newBox(width, height float32) *box {
	var err error

	box := new(box)

	// Sound player

	box.player, err = mandala.NewAudioPlayer()
	if err != nil {
		mandala.Fatalf("%s\n", err.Error())
	}

	// Chipmunk body

	box.physicsShape = chipmunk.NewBox(
		vect.Vect{0, 0},
		vect.Float(width),
		vect.Float(height),
	)

	box.physicsShape.SetElasticity(BoxElasticity)
	box.physicsBody = chipmunk.NewBody(vect.Float(BoxMass), box.physicsShape.Moment(float32(BoxMass)))
	box.physicsBody.AddShape(box.physicsShape)
	box.physicsBody.CallbackHandler = callbacks{}

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
