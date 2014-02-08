package main

import (
	"math"
	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/remogatto/mandala"
	"github.com/remogatto/mathgl"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

const (
	// Y-component for gravity
	Gravity = -900
)

var (
	pyramid = []string{
		"      +            +   ",
		"     +++          +++  ",
		"    +++++         +++  ",
		"   +++++++        +++  ",
		"  +++++++++       +++  ",
	}
)

type world struct {
	width, height                 int
	projMatrix                    mathgl.Mat4f
	viewMatrix                    mathgl.Mat4f
	space                         *chipmunk.Space
	boxes                         []*box
	ground                        *ground
	explosionPlayer               *mandala.AudioPlayer
	explosionBuffer, impactBuffer []byte
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

	// Initialize the audio player
	var err error
	world.explosionPlayer, err = mandala.NewAudioPlayer()
	if err != nil {
		mandala.Fatalf("%s\n", err.Error())
	}

	// Read the PCM audio samples
	responseCh := make(chan mandala.LoadResourceResponse)
	mandala.ReadResource("raw/explosion.pcm", responseCh)
	response := <-responseCh

	if response.Error != nil {
		mandala.Fatalf(response.Error.Error())
	}
	world.explosionBuffer = response.Buffer

	responseCh = make(chan mandala.LoadResourceResponse)
	mandala.ReadResource("raw/impact.pcm", responseCh)
	response = <-responseCh

	if response.Error != nil {
		mandala.Fatalf(response.Error.Error())
	}
	world.impactBuffer = response.Buffer

	return world
}

func (w *world) Projection() mathgl.Mat4f {
	return w.projMatrix
}

func (w *world) View() mathgl.Mat4f {
	return w.viewMatrix
}

func (w *world) createFromString(s []string) {
	// Number of boxes of both axes
	nY := len(s)
	nX := len(s[0])

	// Y coord of the ground
	_, groundY := w.ground.openglShape.Center()
	maxY := float32(w.height)
	maxHeight := float32(maxY) - groundY

	// Calculate box size
	boxW := float32(w.width) / float32(nX)
	boxH := maxHeight / float32(nY)

	// Force a square box
	if boxW >= boxH {
		boxW = boxH
	} else {
		boxH = boxW
	}

	startY := groundY + float32(nY)*boxH

	for y, line := range s {
		for x, b := range line {
			if b == '+' {
				box := newBox(boxW, boxH)
				pos := vect.Vect{
					vect.Float(float32(x) * boxW),
					vect.Float(startY - (float32(y) * boxH)),
				}
				box.physicsBody.SetPosition(pos)
				box.physicsBody.SetAngle(0)
				box.openglShape.Color(colorful.HappyColor())
				w.addBox(box)
			}
		}
	}
}

func (w *world) addBox(box *box) *box {
	box.world = w
	w.space.AddBody(box.physicsBody)
	box.openglShape.AttachToWorld(w)
	w.boxes = append(w.boxes, box)
	box.physicsBody.UserData = impactUserData{box.player, w.impactBuffer}
	return box
}

func (w *world) dropBox(x, y float32) {
	box := newBox(20, 20)
	box.physicsBody.SetMass(200)
	box.physicsBody.AddAngularVelocity(10)
	box.physicsBody.SetAngle(vect.Float(2 * math.Pi * chipmunk.DegreeConst * rand.Float32()))
	box.physicsBody.SetPosition(vect.Vect{vect.Float(x), vect.Float(float32(w.height) - y)})
	w.addBox(box)
}

func (w *world) explosion(x, y float32) {

	w.explosionPlayer.Play(w.explosionBuffer, nil)

	y = float32(w.height) - y
	for _, box := range w.boxes {
		cx, cy := box.openglShape.Center()
		force := vect.Sub(vect.Vect{vect.Float(cx), vect.Float(cy)}, vect.Vect{vect.Float(x), vect.Float(y)})
		length := force.Length()
		force.Normalize()
		force.Mult(vect.Float(1 / length * 3e6))
		box.physicsBody.SetForce(float32(force.X), float32(force.Y))
	}
}

func (w *world) removeBox(box *box, index int) *box {
	box.world = nil
	w.space.RemoveBody(box.physicsBody)
	w.boxes[index] = nil
	w.boxes = append(w.boxes[:index], w.boxes[index+1:]...)
	return box
}

func (w *world) setGround(ground *ground) *ground {
	w.space.AddBody(ground.physicsBody)
	ground.openglShape.AttachToWorld(w)
	w.ground = ground
	return ground
}
