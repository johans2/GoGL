package main

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type renderer struct {
	vao    uint32
	vbo    uint32
	shader uint32
	verts  []float32
}

func (r *renderer) init(verts []float32, program uint32) {
	r.verts = verts
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(r.verts)*4, gl.Ptr(r.verts), gl.STATIC_DRAW)

	// Get the vertex attribute from the shader and point it to data
	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))

	// Get the texCoord attribute from the shader and point it to data
	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))

	// Get the normal attribute from the shader and point it to data
	normalAttrib := uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))
	gl.EnableVertexAttribArray(normalAttrib)
	gl.VertexAttribPointer(normalAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(5*4))
}

func (r *renderer) IssueDrawCall(shader uint32, texture uint32, MVP mgl32.Mat4) {
	// Select the shader to use
	gl.UseProgram(shader)

	// Set the MVP for the object
	MVPuniform := gl.GetUniformLocation(shader, gl.Str("MVP\x00"))
	gl.UniformMatrix4fv(MVPuniform, 1, false, &MVP[0])

	// Bind the vertex array object
	gl.BindVertexArray(r.vao)

	// Bind the created texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Issue drawcall
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(r.verts)))
}
