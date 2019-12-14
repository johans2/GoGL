package main

import (
	"github.com/go-gl/gl/v3.2-core/gl"
)

// TODO: This should clearly use a Uniform Buffer Object.

type globalRenderProperties struct {
	LightDir   [3]float32
	LightColor [3]float32
	CameraPos  [3]float32
	Time       float32
}

// GlobalRenderProps holds the global render properties for all renderers
var GlobalRenderProps globalRenderProperties = globalRenderProperties{LightDir: [3]float32{0.5, 1.2, 1.5}, LightColor: [3]float32{1, 1, 1}}

// ApplyGlobalRenderProperties applies all global render properties to a given shader
func ApplyGlobalRenderProperties(program uint32) {
	gl.UseProgram(program)

	// Apply LightDir
	lightDirUniform := gl.GetUniformLocation(program, gl.Str(lightDirName+"\x00"))
	gl.Uniform3f(lightDirUniform, GlobalRenderProps.LightDir[0], GlobalRenderProps.LightDir[1], GlobalRenderProps.LightDir[2])

	// Apply LightColor
	lightColorUniform := gl.GetUniformLocation(program, gl.Str(lightColorName+"\x00"))
	gl.Uniform3f(lightColorUniform, GlobalRenderProps.LightColor[0], GlobalRenderProps.LightColor[1], GlobalRenderProps.LightColor[2])

	// Apply WorldSpaceCameraPos
	cameraPositionUniform := gl.GetUniformLocation(program, gl.Str(camWorldPosName+"\x00"))
	gl.Uniform3f(cameraPositionUniform, GlobalRenderProps.CameraPos[0], GlobalRenderProps.CameraPos[1], GlobalRenderProps.CameraPos[2])

	// Apply Time
	timeUniform := gl.GetUniformLocation(program, gl.Str(timeName+"\x00"))
	gl.Uniform1f(timeUniform, GlobalRenderProps.Time)
}

// ApplyLightDir applies the light position to a given shader
func ApplyLightDir(program uint32) {
	gl.UseProgram(program)
	lightDirUniform := gl.GetUniformLocation(program, gl.Str(lightDirName+"\x00"))
	gl.Uniform3f(lightDirUniform, GlobalRenderProps.LightDir[0], GlobalRenderProps.LightDir[1], GlobalRenderProps.LightDir[2])
}

// ApplyLightColor applies the light position to a given shader
func ApplyLightColor(program uint32) {
	gl.UseProgram(program)
	lightColorUniform := gl.GetUniformLocation(program, gl.Str(lightColorName+"\x00"))
	gl.Uniform3f(lightColorUniform, GlobalRenderProps.LightColor[0], GlobalRenderProps.LightColor[1], GlobalRenderProps.LightColor[2])
}

// ApplyCameraPosition applies the camera position to a given shader
func ApplyCameraPosition(program uint32) {
	gl.UseProgram(program)
	cameraPositionUniform := gl.GetUniformLocation(program, gl.Str(camWorldPosName+"\x00"))
	gl.Uniform3f(cameraPositionUniform, GlobalRenderProps.CameraPos[0], GlobalRenderProps.CameraPos[1], GlobalRenderProps.CameraPos[2])
}

// ApplyTime applies the time variable to a given shader
func ApplyTime(program uint32) {
	timeUniform := gl.GetUniformLocation(program, gl.Str(timeName+"\x00"))
	gl.Uniform1f(timeUniform, GlobalRenderProps.Time)
}
