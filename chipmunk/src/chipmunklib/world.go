package chipmunklib

import (
	"bytes"
	"image"
	"math"
	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/remogatto/gltext"
	"github.com/remogatto/mandala"
	"github.com/remogatto/mathgl"
	gl "github.com/remogatto/opengles2"
	"github.com/remogatto/shaders"
	"github.com/remogatto/shapes"
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

type texture struct {
	bounds image.Rectangle
	id     uint32
}

func (t *texture) Bounds() image.Rectangle {
	return t.bounds
}

func (t *texture) Id() uint32 {
	return t.id
}

type World struct {
	width, height                 int
	projMatrix                    mathgl.Mat4f
	viewMatrix                    mathgl.Mat4f
	space                         *chipmunk.Space
	boxes                         []*Box
	ground                        *Ground
	explosionPlayer, impactPlayer *mandala.AudioPlayer
	explosionBuffer, impactBuffer []byte
	boxProgramShader              shaders.Program
	segmentProgramShader          shaders.Program
	font                          *gltext.Font
}

func NewWorld(width, height int) *World {
	world := &World{
		width:      width,
		height:     height,
		projMatrix: mathgl.Ortho2D(0, float32(width), 0, float32(height)),
		viewMatrix: mathgl.Ident4f(),
		space:      chipmunk.NewSpace(),
	}

	world.space.Gravity = vect.Vect{0, Gravity}

	// Initialize the audio players
	var err error
	world.explosionPlayer, err = mandala.NewAudioPlayer()
	if err != nil {
		mandala.Fatalf("%s\n", err.Error())
	}

	world.impactPlayer, err = mandala.NewAudioPlayer()
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

	// Compile the shaders

	world.boxProgramShader = shaders.NewProgram(shapes.DefaultBoxFS, shapes.DefaultBoxVS)
	world.segmentProgramShader = shaders.NewProgram(shapes.DefaultSegmentFS, shapes.DefaultSegmentVS)

	// Load the font
	responseCh = make(chan mandala.LoadResourceResponse)
	mandala.ReadResource("raw/freesans.ttf", responseCh)
	response = <-responseCh
	fontBuffer := response.Buffer
	err = response.Error
	if err != nil {
		panic(err)
	}

	world.font, err = gltext.LoadTruetype(bytes.NewBuffer(fontBuffer), world, 12, 32, 127, gltext.LeftToRight)
	if err != nil {
		panic(err)
	}

	return world
}

func (w *World) Projection() mathgl.Mat4f {
	return w.projMatrix
}

func (w *World) View() mathgl.Mat4f {
	return w.viewMatrix
}

func (w *World) UploadRGBAImage(img *image.RGBA) gltext.Texture {
	t := new(texture)
	ib := img.Bounds()
	t.bounds = ib
	gl.GenTextures(1, &t.id)
	gl.BindTexture(gl.TEXTURE_2D, t.id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.Sizei(ib.Dx()), gl.Sizei(ib.Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Void(&img.Pix[0]))
	return t
}

func (w *World) CreateFromString(s []string) {
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
				box := newBox(w, boxW, boxH)
				pos := vect.Vect{
					vect.Float(float32(x) * boxW),
					vect.Float(startY - (float32(y) * boxH)),
				}
				box.physicsBody.SetPosition(pos)
				box.physicsBody.SetAngle(0)
				box.openglShape.SetColor(colorful.HappyColor())
				w.addBox(box)
			}
		}
	}
}

func (w *World) addBox(box *Box) *Box {
	box.world = w
	w.space.AddBody(box.physicsBody)
	box.openglShape.AttachToWorld(w)
	w.boxes = append(w.boxes, box)
	box.physicsBody.UserData = impactUserData{w.impactPlayer, w.impactBuffer}
	return box
}

func (w *World) dropBox(x, y float32) {
	box := newBox(w, 20, 20)
	box.physicsBody.SetMass(10)
	box.physicsBody.AddAngularVelocity(10)
	box.physicsBody.SetAngle(vect.Float(2 * math.Pi * chipmunk.DegreeConst * rand.Float32()))
	box.physicsBody.SetPosition(vect.Vect{vect.Float(x), vect.Float(float32(w.height) - y)})
	w.addBox(box)
}

// Explosion produce an explosion at the given coordinates.
func (w *World) Explosion(x, y float32) {
	w.explosionPlayer.Play(w.explosionBuffer, nil)
	y = float32(w.height) - y
	for _, box := range w.boxes {
		cx, cy := box.openglShape.Center()
		force := vect.Sub(
			vect.Vect{vect.Float(cx / float32(w.width)), vect.Float(cy / float32(w.height))},
			vect.Vect{vect.Float(x / float32(w.width)), vect.Float(y / float32(w.height))},
		)
		force.Normalize()
		force.Mult(vect.Float(1 / force.Length() * 1e5))
		box.physicsBody.SetForce(float32(force.X), float32(force.Y))
	}
}

// Remove removes the box at the given coordinates.
func (w *World) Remove(x, y float32) int {
	y = float32(w.height) - y
	for id, box := range w.boxes {
		cx, cy := box.openglShape.Center()
		distance := vect.Sub(
			vect.Vect{vect.Float(cx), vect.Float(cy)},
			vect.Vect{vect.Float(x), vect.Float(y)},
		)
		if distance.LengthSqr() < vect.Float(BoxSize) {
			w.removeBox(box, id)
			return id
		}
	}
	return -1
}

func (w *World) removeBox(box *Box, index int) {
	box.physicsBody.UserData = nil
	w.space.RemoveBody(box.physicsBody)
	w.boxes = append(w.boxes[:index], w.boxes[index+1:]...)
}

func (w *World) setGround(ground *Ground) *Ground {
	w.space.AddBody(ground.physicsBody)
	ground.openglShape.AttachToWorld(w)
	w.ground = ground
	return ground
}

func (w *World) Destroy() {
	w.impactPlayer.Destroy()
	w.explosionPlayer.Destroy()
	for i := 0; i < len(w.boxes); i++ {
		w.removeBox(w.boxes[i], i)
		i--
	}
}
