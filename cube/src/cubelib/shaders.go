package cubelib

import (
	gl "github.com/remogatto/opengles2"
	"log"
)

type VertexShader string
type FragmentShader string

func checkShaderCompileStatus(shader uint32) {
	var stat int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &stat)
	if stat == 0 {
		var length int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &length)
		infoLog := gl.GetShaderInfoLog(shader, gl.Sizei(length), nil)
		log.Fatalf("Compile error in shader %d: \"%s\"\n", shader, infoLog[:len(infoLog)-1])
	}
}

func checkProgramLinkStatus(pid uint32) {
	var stat int32
	gl.GetProgramiv(pid, gl.LINK_STATUS, &stat)
	if stat == 0 {
		var length int32
		gl.GetProgramiv(pid, gl.INFO_LOG_LENGTH, &length)
		infoLog := gl.GetProgramInfoLog(pid, gl.Sizei(length), nil)
		log.Fatalf("Link error in program %d: \"%s\"\n", pid, infoLog[:len(infoLog)-1])
	}
}

func compileShader(typeOfShader gl.Enum, source string) uint32 {
	if shader := gl.CreateShader(typeOfShader); shader != 0 {
		gl.ShaderSource(shader, 1, &source, nil)
		gl.CompileShader(shader)
		checkShaderCompileStatus(shader)
		return shader
	}
	return 0
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
	p.pid = gl.CreateProgram()
	gl.AttachShader(p.pid, fsh)
	gl.AttachShader(p.pid, vsh)
	gl.LinkProgram(p.pid)
	checkProgramLinkStatus(p.pid)
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
