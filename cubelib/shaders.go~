package cubelib

import (
	gl "github.com/remogatto/opengles2"
	"log"
)

type VertexShader string
type FragmentShader string

func compileShader(typeOfShader gl.Enum, source string) uint32 {
	shader := gl.CreateShader(typeOfShader)
	check()
	gl.ShaderSource(shader, 1, &source, nil)
	check()
	gl.CompileShader(shader)
	check()
	var stat int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &stat)
	if stat == 0 {
		var s = make([]byte, 1000)
		var length gl.Sizei
		_log := string(s)
		gl.GetShaderInfoLog(shader, 1000, &length, &_log)
		log.Fatalf("Error: compiling:\n%s\n", _log)
	}
	return shader
}

func (s VertexShader) Compile() uint32 {
	shaderId := compileShader(gl.VERTEX_SHADER, (string)(s))
	return shaderId
}

func (s FragmentShader) Compile() uint32 {
	shaderId := compileShader(gl.FRAGMENT_SHADER, (string)(s))
	return shaderId
}

// FIXME - it could be of type uint32
type Program struct {
	pid uint32
}

func (p *Program) Link(fsh, vsh uint32) {
	log.Printf("VSH %d FSH %d\n", fsh, vsh)

	p.pid = gl.CreateProgram()
	gl.AttachShader(p.pid, fsh)
	gl.AttachShader(p.pid, vsh)
	gl.LinkProgram(p.pid)
	var stat int32
	gl.GetProgramiv(p.pid, gl.LINK_STATUS, &stat)
	if stat == 0 {
		var s = make([]byte, 1000)
		var length gl.Sizei
		_log := string(s)
		gl.GetProgramInfoLog(p.pid, 1000, &length, &_log)
		log.Fatalf("Error: linking:\n%s\n", _log)
	}
}

func (p *Program) Use() {
	gl.UseProgram(p.pid)
}

func (p *Program) GetAttribute(name string) uint32 {
	return gl.GetAttribLocation(p.pid, name)
}

func (p *Program) GetUniform(name string) uint32 {
	return gl.GetUniformLocation(p.pid, name)
}
