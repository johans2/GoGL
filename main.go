// Copyright 2014 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Renders a textured spinning cube using GLFW 3 and OpenGL 4.1 core forward-compatible profile.
package main

import (
	"bytes"
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

const windowWidth = 1600
const windowHeight = 1200

// Option Example code from nk package
type Option uint8

// Example code form nk package
const (
	Easy Option = 0
	Hard Option = 1
)

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

// State Example code from nk package
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

	var shaderTex shader
	shaderTex.loadFromFile("Assets/simpleTex.vert", "Assets/simpleTex.frag")
	var shaderRed shader
	shaderRed.loadFromFile("Assets/simpleRed.vert", "Assets/simpleRed.frag")
	var shaderGreen shader
	shaderGreen.loadFromFile("Assets/simpleGreen.vert", "Assets/simpleGreen.frag")
	var shaderUnlitColor shader
	shaderUnlitColor.loadFromFile("Assets/unlitColor.vert", "Assets/unlitColor.frag")

	// Set up projection matrix for shader
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10.0)

	// Set up view matrix for shader
	camera := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	// Set up model martix for shader
	model := mgl32.Ident4()

	// Load the model from the obj file
	sphereModel, _ := readOBJ("Assets/lowPolySphere.obj")
	sphereVerts := sphereModel.ToArrayXYZUVN1N2N3()
	boxModel, _ := readOBJ("Assets/box.obj")
	boxVerts := boxModel.ToArrayXYZUVN1N2N3()

	angle := 0.0
	previousTime := glfw.GetTime()

	log.Printf("Finished setup. Now rendering..")

	var bufferVertSource = make([]byte, 1024)
	var bufferFragSource = make([]byte, 1024)

	var activeMaterial material
	activeMaterial.init(&shaderRed)

	var activeModel = sphereVerts

	// Initialize the renderers
	var modelRenderer renderer
	modelRenderer.setData(activeModel, activeMaterial)

	var shaderError error

	for !window.ShouldClose() {
		// Need to reanable these things since Nuklear sets its own gl states when rendering.
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
		MVP := projection.Mul4(camera.Mul4(model))
		modelRenderer.issueDrawCall(MVP)

		// BEGIN GUI
		// Layout GUI
		nk.NkPlatformNewFrame()
		bounds := nk.NkRect(20, 20, 400, 550)
		update := nk.NkBegin(ctxGUI, "Material inspector", bounds,
			nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

		if update > 0 {
			nk.NkLayoutRowStatic(ctxGUI, 10, 80, 1)
			{
				nk.NkLayoutRowDynamic(ctxGUI, 30, 1)
				{
					nk.NkLabel(ctxGUI, "Vertex program:", nk.TextLeft)
					nk.NkEditStringZeroTerminated(ctxGUI, nk.EditField, bufferVertSource, 256, nk.NkFilterDefault)
					nk.NkLabel(ctxGUI, "Fragment program:", nk.TextLeft)
					nk.NkEditStringZeroTerminated(ctxGUI, nk.EditField, bufferFragSource, 256, nk.NkFilterDefault)
					if nk.NkButtonLabel(ctxGUI, "Compile") > 0 {
						var newShader shader
						nVert := bytes.IndexByte(bufferVertSource, 0)
						pathStringVert := string(bufferVertSource[:nVert])
						nFrag := bytes.IndexByte(bufferFragSource, 0)
						pathStringFrag := string(bufferFragSource[:nFrag])

						shaderError = newShader.loadFromFile(pathStringVert, pathStringFrag)
						if shaderError == nil {
							var newMaterial material
							newMaterial.init(&newShader)
							activeMaterial = newMaterial
							modelRenderer.setData(activeModel, activeMaterial)
						}
					}

					if shaderError != nil {
						nk.NkLayoutRowDynamic(ctxGUI, 60, 1)
						{
							log.Printf("ERROR: " + shaderError.Error())
							nk.NkLabelWrap(ctxGUI, "ERROR: "+shaderError.Error())
						}
					}

					nk.NkLabel(ctxGUI, "-------------------------------------", nk.TextCentered)
				}

				modelRenderer.material.drawUI(ctxGUI)

				if len(activeMaterial.fields) != 0 || len(activeMaterial.texBindings) != 0 {
					nk.NkLayoutRowDynamic(ctxGUI, 30, 1)
					{
						if nk.NkButtonLabel(ctxGUI, "Apply") > 0 {
							modelRenderer.material.applyUniforms()
						}
						nk.NkLabel(ctxGUI, "-------------------------------------", nk.TextCentered)
					}
				}

				nk.NkLayoutRowDynamic(ctxGUI, 60, 2)
				{
					if nk.NkButtonLabel(ctxGUI, "Sphere") > 0 {
						activeModel = sphereVerts
						modelRenderer.setData(activeModel, activeMaterial)
						modelRenderer.material.applyUniforms()
					}
					if nk.NkButtonLabel(ctxGUI, "Box") > 0 {
						activeModel = boxVerts
						modelRenderer.setData(activeModel, activeMaterial)
						modelRenderer.material.applyUniforms()
					}

				}

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
	glfw.Terminate()
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
