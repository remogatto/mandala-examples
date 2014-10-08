package chipmunklib

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

type Box struct {
	// Chipumunk stuff
	physicsBody  *chipmunk.Body
	physicsShape *chipmunk.Shape

	// OpenGL stuff
	openglShape *shapes.Box

	world *World
}

func (c callbacks) CollisionEnter(arbiter *chipmunk.Arbiter) bool {
	a := arbiter.BodyA
	b := arbiter.BodyB
	impact, ok := a.UserData.(impactUserData)
	if ok {
		impact.player.Play(impact.impactBuffer, nil)
	}
	impact, ok = b.UserData.(impactUserData)
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

func newBox(world *World, width, height float32) *Box {
	box := new(Box)

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

	box.openglShape = shapes.NewBox(world.boxProgramShader, width, height)

	return box
}

func (box *Box) draw() {
	pos := box.physicsBody.Position()
	rot := box.physicsBody.Angle() * chipmunk.DegreeConst
	box.openglShape.MoveTo(float32(pos.X), float32(pos.Y))
	box.openglShape.Rotate(float32(rot))
	box.openglShape.Draw()
}

func (box *Box) inViewport() bool {
	pos := box.physicsBody.Position()
	r := box.openglShape.Bounds()
	width := float32(r.Dx())
	return (float32(pos.X) > -width) && (float32(pos.X) < (width + float32(box.world.width)))
}
