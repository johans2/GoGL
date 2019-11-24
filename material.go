package main

import (
	"bytes"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/golang-ui/nuklear/nk"
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
	draw(glContext *nk.Context)
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
			m.fields = append(m.fields, &matFieldTexture{uniform.name, tex, make([]byte, 1024)})
		}

	}
}

func (m *material) drawUI(glContext *nk.Context) {
	for _, field := range m.fields {
		field.draw(glContext)
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

func (f *matFieldFloat) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 2)
	{
		nk.NkLabel(glContext, f.name, nk.TextLeft)
		nk.NkPropertyFloat(glContext, "value: ", -999, &f.value, 999, 0.01, 0.1)
	}
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

func (v2 *matFieldVec2) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 3)
	{
		nk.NkLabel(glContext, v2.name, nk.TextLeft)
		nk.NkPropertyFloat(glContext, "x: ", -999, &v2.x, 999, 0.01, 0.1)
		nk.NkPropertyFloat(glContext, "y: ", -999, &v2.y, 999, 0.01, 0.1)
	}
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

func (v3 *matFieldVec3) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 4)
	{
		nk.NkLabel(glContext, v3.name, nk.TextLeft)
		nk.NkPropertyFloat(glContext, "x: ", -999, &v3.x, 999, 0.01, 0.1)
		nk.NkPropertyFloat(glContext, "y: ", -999, &v3.y, 999, 0.01, 0.1)
		nk.NkPropertyFloat(glContext, "z: ", -999, &v3.z, 999, 0.01, 0.1)
	}
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

func (v4 *matFieldVec4) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 5)
	{
		nk.NkLabel(glContext, v4.name, nk.TextLeft)
		nk.NkPropertyFloat(glContext, "x: ", -999, &v4.x, 999, 0.01, 0.1)
		nk.NkPropertyFloat(glContext, "y: ", -999, &v4.y, 999, 0.01, 0.1)
		nk.NkPropertyFloat(glContext, "z: ", -999, &v4.z, 999, 0.01, 0.1)
		nk.NkPropertyFloat(glContext, "w: ", -999, &v4.w, 999, 0.01, 0.1)
	}
}

func (v4 *matFieldVec4) apply(mat *material) {
	uniform := gl.GetUniformLocation(mat.shader.program, gl.Str("material."+v4.name+"\x00"))
	gl.Uniform4f(uniform, v4.x, v4.y, v4.z, v4.w)
}

// Texture
type matFieldTexture struct {
	name     string
	tex      texture
	filePath []byte
}

func (t *matFieldTexture) draw(glContext *nk.Context) {
	nk.NkLayoutRowDynamic(glContext, 30, 1)
	{
		nk.NkLabel(glContext, t.name+": ", nk.TextLeft)
		nk.NkEditStringZeroTerminated(glContext, nk.EditField, t.filePath, 1024, nk.NkFilterDefault)
	}
}

func (t *matFieldTexture) apply(mat *material) {
	// Load it from file
	n := bytes.IndexByte(t.filePath, 0)
	pathString := string(t.filePath[:n])
	t.tex.loadFromFile(pathString)

	// Get the uniform location
	uniform := gl.GetUniformLocation(mat.shader.program, gl.Str("material."+t.name+"\x00"))

	// Create and add a new texture binding struct
	var texBind textureBinding
	texBind.glTexID = t.tex.id
	texBind.uniformLocation = uniform
	mat.texBindings = append(mat.texBindings, texBind)
}
