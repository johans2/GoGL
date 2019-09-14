// Copyright 2014 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Renders a textured spinning cube using GLFW 3 and OpenGL 4.1 core forward-compatible profile.
package main

import (
	"fmt"
	"go/build"
	_ "image/png"
	"log"
	"os"
	"runtime"

	"github.com/go-gl/gl/v3.2-core/gl"
	//"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
)

const windowWidth = 800
const windowHeight = 600

type Option uint8

const (
	Easy Option = 0
	Hard Option = 1
)

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

type State struct {
	bgColor nk.Color
	prop    int32
	opt     Option
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Init nuklear gui
	log.Printf("NkPlatformInit")
	ctxGUI := nk.NkPlatformInit(window, nk.PlatformInstallCallbacks)

	// Create font
	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	//sansFont := nk.NkFontAtlasAddFromBytes(atlas, MustAsset("assets/FreeSans.ttf"), 16, nil)
	sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(ctxGUI, sansFont.Handle())
	}
	log.Println("Finished setting up Nk-GUI")

	state := &State{
		bgColor: nk.NkRgba(28, 48, 62, 255),
	}

	window.MakeContextCurrent()

	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	// Set up projection matrix for shader
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	// Set up view matrix for shader
	camera := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	// Set up model martix for shader
	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	// Set up texture for shader
	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Load the texture
	texture, err := newTexture("Assets/square.png")
	if err != nil {
		log.Fatalln(err)
	}

	// Load the model from the obj file
	sphereModel, _ := readOBJ("Assets/lowPolySphere.obj")
	sphereVerts := sphereModel.ToArrayXYZUVN1N2N3()
	boxModel, _ := readOBJ("Assets/box.obj")
	boxVerts := boxModel.ToArrayXYZUVN1N2N3()

	// Configure the vertex data 1-------------------------------
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(sphereVerts)*4, gl.Ptr(sphereVerts), gl.STATIC_DRAW)

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

	// Configure the vertex data 2-----------------------------
	var vao2 uint32
	gl.GenVertexArrays(1, &vao2)
	gl.BindVertexArray(vao2)

	var vbo2 uint32
	gl.GenBuffers(1, &vbo2)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo2)
	gl.BufferData(gl.ARRAY_BUFFER, len(boxVerts)*4, gl.Ptr(boxVerts), gl.STATIC_DRAW)

	// Get the vertex attribute from the shader and point it to data
	vertAttrib2 := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib2)
	gl.VertexAttribPointer(vertAttrib2, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))

	// Get the texCoord attribute from the shader and point it to data
	texCoordAttrib2 := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib2)
	gl.VertexAttribPointer(texCoordAttrib2, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))

	// Get the normal attribute from the shader and point it to data
	normalAttrib2 := uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))
	gl.EnableVertexAttribArray(normalAttrib2)
	gl.VertexAttribPointer(normalAttrib2, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(5*4))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	angle := 0.0
	previousTime := glfw.GetTime()

	log.Printf("Finished setup. Now rendering..")

	var activeVAO uint32 = vao
	var activeLen int32 = int32(len(sphereVerts))
	var buffer []byte = make([]byte, 256)

	for !window.ShouldClose() {
		//window.MakeContextCurrent()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.CULL_FACE)
		gl.DepthFunc(gl.LESS)
		gl.ClearColor(1.0, 1.0, 1.0, 1.0)

		// Update
		rotSpeed := 0.5
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += (elapsed * rotSpeed)
		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})

		// Render
		gl.UseProgram(program)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		gl.BindVertexArray(activeVAO)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.DrawArrays(gl.TRIANGLES, 0, activeLen)

		// BEGIN GUI
		// Layout GUI
		nk.NkPlatformNewFrame()
		bounds := nk.NkRect(50, 50, 230, 250)
		update := nk.NkBegin(ctxGUI, "Demo", bounds,
			nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

		if update > 0 {
			nk.NkLayoutRowStatic(ctxGUI, 30, 80, 1)
			{
				if nk.NkButtonLabel(ctxGUI, "button") > 0 {
					activeVAO = vao
					activeLen = int32(len(sphereVerts))
				}
				if nk.NkButtonLabel(ctxGUI, "button2") > 0 {
					activeVAO = vao2
					activeLen = int32(len(boxVerts))
				}
				if nk.NkButtonLabel(ctxGUI, "button") > 0 {
					log.Println("[INFO] button pressed!")
				}
				if nk.NkButtonLabel(ctxGUI, "button") > 0 {
					log.Println("[INFO] button pressed!")
				}

				nk.NkEditStringZeroTerminated(ctxGUI, nk.EditField,
					buffer, 256, nk.NkFilterDefault)
			}
		}
		nk.NkEnd(ctxGUI)

		// Render GUI
		bg := make([]float32, 4)
		nk.NkColorFv(bg, state.bgColor)
		width, height := window.GetSize()
		gl.Viewport(0, 0, int32(width), int32(height))
		nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
		// END GUI

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}

	//nk.NkPlatformShutdown()
	//glfw.Terminate()
}

func createAndBindVAO(id *uint32, verts []float32) {

}

// Set the working directory to the root of Go package, so that its assets can be accessed.
func init() {
	dir, err := importPathToDir("GoGL")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}
}

// importPathToDir resolves the absolute path from importPath.
// There doesn't need to be a valid Go package inside that import path,
// but the directory must exist.
func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}
