package main

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type renderer struct {
	vao      uint32
	vbo      uint32
	verts    []float32
	material material
}

func (r *renderer) setData(verts []float32, material material) {
	r.verts = verts
	r.material = material
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(r.verts)*4, gl.Ptr(r.verts), gl.STATIC_DRAW)

	// Get the vertex attribute from the shader and point it to data
	vertAttrib := uint32(gl.GetAttribLocation(r.material.shader.program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))

	// Get the texCoord attribute from the shader and point it to data
	texCoordAttrib := uint32(gl.GetAttribLocation(r.material.shader.program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))

	// Get the normal attribute from the shader and point it to data
	normalAttrib := uint32(gl.GetAttribLocation(r.material.shader.program, gl.Str("normal\x00")))
	gl.EnableVertexAttribArray(normalAttrib)
	gl.VertexAttribPointer(normalAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(5*4))

	gl.BindFragDataLocation(r.material.shader.program, 0, gl.Str("outputColor\x00"))
}

func (r *renderer) issueDrawCall(model mgl32.Mat4, view mgl32.Mat4, projection mgl32.Mat4, cameraWorldPos mgl32.Vec3) {
	// Select the shader to use
	gl.UseProgram(r.material.shader.program)

	// Set the modelUniform for the object
	modelUniform := gl.GetUniformLocation(r.material.shader.program, gl.Str(modelMatrixName+"\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	// Set the viewUniform for the object
	viewUniform := gl.GetUniformLocation(r.material.shader.program, gl.Str(viewMatrixName+"\x00"))
	gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])

	// Set the projectionUniform for the object
	projectionUniform := gl.GetUniformLocation(r.material.shader.program, gl.Str(projMatrixName+"\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	// Also pass the combined MVP uniform for convenience
	MVP := projection.Mul4(view.Mul4(model))
	MVPUniform := gl.GetUniformLocation(r.material.shader.program, gl.Str(mvpMatrixName+"\x00"))
	gl.UniformMatrix4fv(MVPUniform, 1, false, &MVP[0])

	camWorldPosUniform := gl.GetUniformLocation(r.material.shader.program, gl.Str(camWorldPosName+"\x00"))
	gl.Uniform3f(camWorldPosUniform, cameraWorldPos.X(), cameraWorldPos.Y(), cameraWorldPos.Z())

	// Bind the vertex array object
	gl.BindVertexArray(r.vao)

	// Bind the material textures
	r.material.bindTextures()

	// Issue drawcall
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(r.verts)))
}
