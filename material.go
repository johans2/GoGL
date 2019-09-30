package main

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/golang-ui/nuklear/nk"
)

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
		}

	}
}

func (m *material) drawUI(glContext *nk.Context) {
	for _, field := range m.fields {
		field.draw(glContext)
	}
}

func (m *material) applyValues() {
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

func (f *matFieldVec2) draw(glContext *nk.Context) {
	nk.NkPropertyFloat(glContext, "x: ", -9999, &f.x, 9999, 0.1, 0.01)
	nk.NkPropertyFloat(glContext, "y: ", -9999, &f.y, 9999, 0.1, 0.01)
}

func (f *matFieldVec2) apply(shader *shader) {
	uniform := gl.GetUniformLocation(shader.program, gl.Str(f.name+"\x00"))
	gl.Uniform2f(uniform, f.x, f.y)
}

// Vec3
type matFieldVec3 struct {
	name string
	x    float32
	y    float32
	z    float32
}

func (f *matFieldVec3) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 4)
	{
		nk.NkLabel(glContext, f.name, nk.TextLeft)
		nk.NkPropertyFloat(glContext, "x: ", -9999, &f.x, 9999, 1, 1)
		nk.NkPropertyFloat(glContext, "y: ", -9999, &f.y, 9999, 1, 1)
		nk.NkPropertyFloat(glContext, "z: ", -9999, &f.z, 9999, 1, 1)
	}
}

func (f *matFieldVec3) apply(shader *shader) {
	uniform := gl.GetUniformLocation(shader.program, gl.Str(f.name+"\x00"))
	gl.Uniform3f(uniform, f.x, f.y, f.z)
}
