package cubelib

import (
	"fmt"
	"github.com/mortdeus/mathgl"
	gl "github.com/remogatto/opengles2"
	"image"
	"image/png"
	"os"
)

const (
	SizeOfFloat   = 4
	TEX_COORD_MAX = 1
)

type buffer struct {
	target uint32
}

func (b *buffer) bind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, b.target)
}

type BufferByte struct {
	data []byte
	buffer
}

func (b *BufferByte) Len() int {
	return len(b.data)
}

func (b *BufferByte) Data() []byte {
	return b.data
}

type BufferFloat struct {
	data []float32
	buffer
}

func (b *BufferFloat) Len() int {
	return len(b.data)
}

func (b *BufferFloat) Data() []float32 {
	return b.data
}

func NewBufferByte(data []byte) *BufferByte {
	b := new(BufferByte)
	b.data = data
	gl.GenBuffers(1, &b.target)
	b.bind()
	gl.BufferData(gl.ARRAY_BUFFER, gl.SizeiPtr(len(b.data)), gl.Void(&b.data[0]), gl.STATIC_DRAW)
	return b
}

func NewBufferFloat(data []float32) *BufferFloat {
	b := new(BufferFloat)
	b.data = data
	gl.GenBuffers(1, &b.target)
	b.bind()
	gl.BufferData(gl.ARRAY_BUFFER, gl.SizeiPtr(len(b.data))*SizeOfFloat, gl.Void(&b.data[0]), gl.STATIC_DRAW)
	return b
}

func check() {
	error := gl.GetError()
	if error != 0 {
		panic(fmt.Sprintf("An error occurred! Code: 0x%x", error))
	}
}

type Cube struct {
	Vertices                      *BufferFloat
	Program                       Program
	indices                       *BufferByte
	textureBuffer                 uint32
	attrPos, attrColor, attrTexIn uint32
	uniformTexture                uint32
	uniformModel                  uint32
	uniformProjectionView         uint32
	model, projectionView         mathgl.Mat4f
}

func NewCube() *Cube {
	cube := new(Cube)
	cube.model = mathgl.Ident4f()

	cube.Vertices = NewBufferFloat([]float32{
		// Front
		1, -1, 1, 1, TEX_COORD_MAX, 0,
		1, 1, 1, 1, TEX_COORD_MAX, TEX_COORD_MAX,
		-1, 1, 1, 1, 0, TEX_COORD_MAX,
		-1, -1, 1, 1, 0, 0,
		// Back
		1, 1, -1, 1, TEX_COORD_MAX, 0,
		-1, -1, -1, 1, 0, TEX_COORD_MAX,
		1, -1, -1, 1, TEX_COORD_MAX, TEX_COORD_MAX,
		-1, 1, -1, 1, 0, 0,
		// Left
		-1, -1, 1, 1, TEX_COORD_MAX, 0,
		-1, 1, 1, 1, TEX_COORD_MAX, TEX_COORD_MAX,
		-1, 1, -1, 1, 0, TEX_COORD_MAX,
		-1, -1, -1, 1, 0, 0,
		// Right
		1, -1, -1, 1, TEX_COORD_MAX, 0,
		1, 1, -1, 1, TEX_COORD_MAX, TEX_COORD_MAX,
		1, 1, 1, 1, 0, TEX_COORD_MAX,
		1, -1, 1, 1, 0, 0,
		// Top
		1, 1, 1, 1, TEX_COORD_MAX, 0,
		1, 1, -1, 1, TEX_COORD_MAX, TEX_COORD_MAX,
		-1, 1, -1, 1, 0, TEX_COORD_MAX,
		-1, 1, 1, 1, 0, 0,
		// Bottom
		1, -1, -1, 1, TEX_COORD_MAX, 0,
		1, -1, 1, 1, TEX_COORD_MAX, TEX_COORD_MAX,
		-1, -1, 1, 1, 0, TEX_COORD_MAX,
		-1, -1, -1, 1, 0, 0,
	})
	cube.indices = NewBufferByte([]byte{
		// Front
		0, 1, 2,
		2, 3, 0,
		// Back
		4, 5, 6,
		4, 5, 7,
		// Left
		8, 9, 10,
		10, 11, 8,
		// Right
		12, 13, 14,
		14, 15, 12,
		// Top
		16, 17, 18,
		18, 19, 16,
		// Bottom
		20, 21, 22,
		22, 23, 20,
	})

	fragmentShader := (FragmentShader)(`
        precision highp float;
	varying vec2 texOut;
        uniform sampler2D texture;

	void main() {
		gl_FragColor = texture2D(texture, texOut);
	}
        `)
	vertexShader := (VertexShader)(`
        uniform mat4 model;
        uniform mat4 projection_view;
        attribute vec4 pos;
        attribute vec2 texIn;
        varying vec2 texOut;

        void main() {
          gl_Position = projection_view*model*pos;
          texOut = texIn;
        }
        `)

	fsh := fragmentShader.Compile()
	vsh := vertexShader.Compile()
	cube.Program.Link(fsh, vsh)

	cube.Program.Use()
	cube.attrPos = cube.Program.GetAttribute("pos")
	cube.attrColor = cube.Program.GetAttribute("color")
	cube.attrTexIn = cube.Program.GetAttribute("texIn")
	cube.uniformTexture = cube.Program.GetUniform("texture")

	cube.uniformModel = cube.Program.GetUniform("model")
	cube.uniformProjectionView = cube.Program.GetUniform("projection_view")

	gl.EnableVertexAttribArray(cube.attrPos)
	gl.EnableVertexAttribArray(cube.attrColor)
	gl.EnableVertexAttribArray(cube.attrTexIn)

	return cube
}

func (c *Cube) Rotate(angle float32, axis mathgl.Vec3f) {
	c.model = mathgl.HomogRotate3D(angle, axis)
}

func (c *Cube) AttachTextureFromFile(filename string) error {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	// Decode the image.
	img, err := png.Decode(file)
	if err != nil {
		return err
	}
	c.AttachTexture(img)
	return nil
}

func (c *Cube) AttachTexture(img image.Image) {
	bounds := img.Bounds()
	width, height := bounds.Size().X, bounds.Size().Y
	buffer := make([]byte, width*height*4)
	index := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			buffer[index] = byte(r)
			buffer[index+1] = byte(g)
			buffer[index+2] = byte(b)
			buffer[index+3] = byte(a)
			index += 4
		}
	}
	gl.GenTextures(1, &c.textureBuffer)
	gl.BindTexture(gl.TEXTURE_2D, c.textureBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.Sizei(width), gl.Sizei(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Void(&buffer[0]))
}

func (c *Cube) AttachTextureFromBuffer(buffer []byte, width, height int) {
	gl.GenTextures(1, &c.textureBuffer)
	gl.BindTexture(gl.TEXTURE_2D, c.textureBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.Sizei(width), gl.Sizei(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Void(&buffer[0]))
}

func (c *Cube) Draw() {
	c.Program.Use()
	c.Vertices.bind()
	gl.VertexAttribPointer(c.attrPos, 4, gl.FLOAT, false, SizeOfFloat*6, 0)
	gl.VertexAttribPointer(c.attrTexIn, 2, gl.FLOAT, false, 6*SizeOfFloat, 4*SizeOfFloat)
	//	gl.VertexAttribPointer(c.attrColor, 4, gl.FLOAT, false, SizeOfFloat*8, SizeOfFloat*4)

	gl.UniformMatrix4fv(int32(c.uniformModel), 1, false, (*float32)(&c.model[0]))
	gl.UniformMatrix4fv(int32(c.uniformProjectionView), 1, false, (*float32)(&c.projectionView[0]))

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, c.textureBuffer)
	gl.Uniform1i(int32(c.uniformTexture), 0)

	gl.DrawElements(gl.TRIANGLES, gl.Sizei(c.indices.Len()), gl.UNSIGNED_BYTE, gl.Void(&c.indices.Data()[0]))
	gl.Flush()
	gl.Finish()
}
