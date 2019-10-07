package main

import (
	"bytes"
	"log"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/golang-ui/nuklear/nk"
)

var texUnit int32

type material struct {
	shader *shader
	fields []materialField
}

type materialField interface {
	draw(glContext *nk.Context)
	apply(shader *shader)
}

// Material functions
func (m *material) Init(shader *shader) {
	m.shader = shader

	for _, uniform := range shader.uniforms {
		switch uniform.uType {
		case uniformFloat:
			m.fields = append(m.fields, &matFieldFloat{uniform.name, 0})
		case uniformVec2:
			m.fields = append(m.fields, &matFieldVec2{uniform.name, 0, 0})
		case uniformVec3:
			m.fields = append(m.fields, &matFieldVec3{uniform.name, 1, 0, 0})
		case uniformTex2D:
			tex := texture{}
			m.fields = append(m.fields, &matFieldTexture{uniform.name, tex, make([]byte, 1024)})
		}

	}
}

func (m *material) drawUI(glContext *nk.Context) {
	for _, field := range m.fields {
		field.draw(glContext)
	}
}

func (m *material) applyValues() {
	texUnit = 0
	for _, field := range m.fields {
		field.apply(m.shader)
	}
}

// Field implementations and functions

// Float
type matFieldFloat struct {
	name  string
	value float32
}

func (f *matFieldFloat) draw(glContext *nk.Context) {
	nk.NkPropertyFloat(glContext, "value: ", -9999, &f.value, 9999, 0.1, 0.01)
}

func (f *matFieldFloat) apply(shader *shader) {
	uniform := gl.GetUniformLocation(shader.program, gl.Str(f.name+"\x00"))
	gl.Uniform1f(uniform, f.value)
}

// Vec2
type matFieldVec2 struct {
	name string
	x    float32
	y    float32
}

func (v2 *matFieldVec2) draw(glContext *nk.Context) {
	nk.NkPropertyFloat(glContext, "x: ", -9999, &v2.x, 9999, 0.1, 0.01)
	nk.NkPropertyFloat(glContext, "y: ", -9999, &v2.y, 9999, 0.1, 0.01)
}

func (v2 *matFieldVec2) apply(shader *shader) {
	uniform := gl.GetUniformLocation(shader.program, gl.Str(v2.name+"\x00"))
	gl.Uniform2f(uniform, v2.x, v2.y)
}

// Vec3
type matFieldVec3 struct {
	name string
	x    float32
	y    float32
	z    float32
}

func (v3 *matFieldVec3) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 4)
	{
		nk.NkLabel(glContext, v3.name, nk.TextLeft)
		nk.NkPropertyFloat(glContext, "x: ", -9999, &v3.x, 9999, 1, 1)
		nk.NkPropertyFloat(glContext, "y: ", -9999, &v3.y, 9999, 1, 1)
		nk.NkPropertyFloat(glContext, "z: ", -9999, &v3.z, 9999, 1, 1)
	}
}

func (v3 *matFieldVec3) apply(shader *shader) {
	uniform := gl.GetUniformLocation(shader.program, gl.Str(v3.name+"\x00"))
	gl.Uniform3f(uniform, v3.x, v3.y, v3.z)
}

// Vec4
type matFieldVec4 struct {
	name string
	x    float32
	y    float32
	z    float32
	w    float32
}

func (v4 *matFieldVec4) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 4)
	{
		nk.NkLabel(glContext, v4.name, nk.TextLeft)
		nk.NkPropertyFloat(glContext, "x: ", -9999, &v4.x, 9999, 1, 1)
		nk.NkPropertyFloat(glContext, "y: ", -9999, &v4.y, 9999, 1, 1)
		nk.NkPropertyFloat(glContext, "z: ", -9999, &v4.z, 9999, 1, 1)
		nk.NkPropertyFloat(glContext, "w: ", -9999, &v4.w, 9999, 1, 1)
	}
}

func (v4 *matFieldVec4) apply(shader *shader) {
	uniform := gl.GetUniformLocation(shader.program, gl.Str(v4.name+"\x00"))
	gl.Uniform4f(uniform, v4.x, v4.y, v4.z, v4.w)
}

type matFieldTexture struct {
	name     string
	tex      texture
	filePath []byte
}

func (t *matFieldTexture) draw(glContext *nk.Context) {
	//t.filePath = make([]byte, 1024)
	nk.NkEditStringZeroTerminated(glContext, nk.EditField, t.filePath, 1024, nk.NkFilterDefault)
}

func (t *matFieldTexture) apply(shader *shader) {
	// Load it from file
	n := bytes.IndexByte(t.filePath, 0)
	pathString := string(t.filePath[:n])
	t.tex.loadFromFile(pathString)

	log.Printf(pathString)

	// Get the uniform location
	uniform := gl.GetUniformLocation(shader.program, gl.Str(t.name+"\x00"))
	// Bind the uniforms to texture units/channels. e.g..
	/*
		glUniform1i(decalTexLocation, 0);
		glUniform1i(bumpTexLocation,  1);
	*/
	gl.Uniform1i(uniform, texUnit)
	// Get next available texture unit 0...1....2....3...etc
	gl.ActiveTexture(gl.TEXTURE0 + uint32(texUnit))
	gl.BindTexture(gl.TEXTURE_2D, t.tex.id)

	texUnit++
}
