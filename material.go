package main

import (
	"fmt"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/inkyblackness/imgui-go"
)

var texUnit int32

type material struct {
	shader      shader
	fields      []materialField
	texBindings []textureBinding
}

type textureBinding struct {
	glTexID         uint32
	uniformLocation int32
}

type materialField interface {
	draw()
	apply(mat *material)
}

// Material functions
func (m *material) init(shader shader) {
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
			m.fields = append(m.fields, &matFieldTexture{uniform.name, tex, ""})
		}

	}
}

func (m *material) drawUI() {
	for _, field := range m.fields {
		field.draw()
	}
}

func (m *material) applyUniforms() {
	texUnit = 0
	// Clear texturebinding list
	for _, field := range m.fields {
		field.apply(m)
	}
}

func (m *material) bindTextures() {
	texUnit := uint32(0)
	for _, texBinding := range m.texBindings {
		// Set the texture uniform value
		gl.Uniform1i(texBinding.uniformLocation, int32(texUnit))

		gl.ActiveTexture(gl.TEXTURE0 + texUnit)
		gl.BindTexture(gl.TEXTURE_2D, texBinding.glTexID)
		texUnit++
	}
}

// Field implementations and functions

// Float
type matFieldFloat struct {
	name  string
	value float32
}

func (f *matFieldFloat) draw() {
	imgui.Text(f.name)
	imgui.SameLine()
	imgui.DragFloat(f.name, &f.value)
}

func (f *matFieldFloat) apply(mat *material) {
	uniform := gl.GetUniformLocation(mat.shader.program, gl.Str("material."+f.name+"\x00"))
	gl.Uniform1f(uniform, f.value)
}

// Vec2
type matFieldVec2 struct {
	name string
	x    float32
	y    float32
}

func (v2 *matFieldVec2) draw() {
	imgui.Text(v2.name)
	imgui.SameLine()
	imgui.DragFloat("x", &v2.x)
	imgui.SameLine()
	imgui.DragFloat("y", &v2.y)
}

func (v2 *matFieldVec2) apply(mat *material) {
	uniform := gl.GetUniformLocation(mat.shader.program, gl.Str("material."+v2.name+"\x00"))
	gl.Uniform2f(uniform, v2.x, v2.y)
}

// Vec3
type matFieldVec3 struct {
	name string
	x    float32
	y    float32
	z    float32
}

func (v3 *matFieldVec3) draw() {
	imgui.Text(v3.name)
	imgui.SameLine()
	imgui.DragFloat("x", &v3.x)
	imgui.SameLine()
	imgui.DragFloat("y", &v3.y)
	imgui.SameLine()
	imgui.DragFloat("z", &v3.z)
}

func (v3 *matFieldVec3) apply(mat *material) {
	uniform := gl.GetUniformLocation(mat.shader.program, gl.Str("material."+v3.name+"\x00"))
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

func (v4 *matFieldVec4) draw() {
	imgui.Text(v4.name)
	imgui.SameLine()
	imgui.DragFloat("x", &v4.x)
	imgui.SameLine()
	imgui.DragFloat("y", &v4.y)
	imgui.SameLine()
	imgui.DragFloat("z", &v4.z)
	imgui.SameLine()
	imgui.DragFloat("w", &v4.w)
}

func (v4 *matFieldVec4) apply(mat *material) {
	uniform := gl.GetUniformLocation(mat.shader.program, gl.Str("material."+v4.name+"\x00"))
	gl.Uniform4f(uniform, v4.x, v4.y, v4.z, v4.w)
}

// Texture
type matFieldTexture struct {
	name     string
	tex      texture
	filePath string
}

func (t *matFieldTexture) draw() {
	imgui.Text(t.name)
	imgui.SameLine()
	imgui.InputText("", &t.filePath)
}

func (t *matFieldTexture) apply(mat *material) {
	texError := t.tex.loadFromFile(t.filePath)

	if texError != nil {
		fmt.Println("Bad texture" + texError.Error())
	}

	// Get the uniform location
	uniform := gl.GetUniformLocation(mat.shader.program, gl.Str("material."+t.name+"\x00"))

	// Create and add a new texture binding struct
	var texBind textureBinding
	texBind.glTexID = t.tex.id
	texBind.uniformLocation = uniform
	mat.texBindings = append(mat.texBindings, texBind)
}
