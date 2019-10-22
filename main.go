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

type mouseState bool

const (
	press   mouseState = true
	release mouseState = false
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

func onMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButton1 && action == glfw.Press {
		log.Println("MB 1 Press!")
	}

	if button == glfw.MouseButton1 && action == glfw.Release {
		log.Println("MB 1 Release!")
	}

}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	currentMouseState := release
	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		if button == glfw.MouseButton1 && action == glfw.Press {
			currentMouseState = press
		}

		if button == glfw.MouseButton1 && action == glfw.Release {
			currentMouseState = release
		}
	})

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
		bgColor: nk.NkRgba(255, 255, 255, 255),
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
	view := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	// Set up model martix for shader
	model := mgl32.Ident4()

	// Load the model from the obj file
	sphereModel, _ := readOBJ("Assets/sphere.obj")
	sphereVerts := sphereModel.ToArrayXYZUVN1N2N3()
	boxModel, _ := readOBJ("Assets/box.obj")
	boxVerts := boxModel.ToArrayXYZUVN1N2N3()
	torusModel, _ := readOBJ("Assets/torus.obj")
	torusVerts := torusModel.ToArrayXYZUVN1N2N3()
	planeModel, _ := readOBJ("Assets/plane.obj")
	planeVerts := planeModel.ToArrayXYZUVN1N2N3()
	coneModel, _ := readOBJ("Assets/cone.obj")
	coneVerts := coneModel.ToArrayXYZUVN1N2N3()

	angle := 0.0
	previousTime := glfw.GetTime()

	log.Printf("Finished setup. Now rendering..")

	var bufferVertSource = make([]byte, 1024)
	var bufferFragSource = make([]byte, 1024)

	var activeMaterial material
	activeMaterial.init(&shaderGreen)

	var activeModel = sphereVerts

	// Initialize the renderers
	var modelRenderer renderer
	modelRenderer.setData(activeModel, activeMaterial)

	var shaderError error
	clearColor := mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	rotationSpeed := float32(0.5)
	scale := float32(1.0)
	var mouseXPrev float64
	var mouseYPrev float64
	var mouseY float64
	var mouseX float64

	for !window.ShouldClose() {

		// Need to reanable these things since Nuklear sets its own gl states when rendering.
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.CULL_FACE)
		gl.DepthFunc(gl.LESS)
		gl.ClearColor(clearColor.X(), clearColor.Y(), clearColor.Z(), clearColor.W())

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		if currentMouseState == press {
			x, y := window.GetCursorPos()
			mouseX = float64(x)
			mouseY = y
			deltaX := mouseX - mouseXPrev
			deltaY := mouseY - mouseYPrev
			angle += (deltaX/100 + deltaY/100)
			mouseXPrev = mouseX
			mouseYPrev = mouseY
		} else {
			angle += (elapsed * float64(rotationSpeed))
		}

		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		model = model.Mul4(mgl32.Scale3D(scale, scale, scale))

		// Render
		modelRenderer.issueDrawCall(model, view, projection)

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
					nk.NkLabel(ctxGUI, "SHADER FILES", nk.TextCentered)
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
				}

				if len(activeMaterial.fields) != 0 || len(activeMaterial.texBindings) != 0 {
					nk.NkLayoutRowDynamic(ctxGUI, 30, 1)
					{
						nk.NkLabel(ctxGUI, "SHADER PROPERTIES", nk.TextCentered)
						modelRenderer.material.drawUI(ctxGUI)
						if nk.NkButtonLabel(ctxGUI, "Apply") > 0 {
							modelRenderer.material.applyUniforms()
						}
					}
				}
				nk.NkLabel(ctxGUI, "", nk.TextCentered)
				nk.NkLabel(ctxGUI, "MODEL", nk.TextCentered)
				nk.NkLayoutRowDynamic(ctxGUI, 60, 5)
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
					if nk.NkButtonLabel(ctxGUI, "Torus") > 0 {
						activeModel = torusVerts
						modelRenderer.setData(activeModel, activeMaterial)
						modelRenderer.material.applyUniforms()
					}
					if nk.NkButtonLabel(ctxGUI, "Plane") > 0 {
						activeModel = planeVerts
						modelRenderer.setData(activeModel, activeMaterial)
						modelRenderer.material.applyUniforms()
					}
					if nk.NkButtonLabel(ctxGUI, "Cone") > 0 {
						activeModel = coneVerts
						modelRenderer.setData(activeModel, activeMaterial)
						modelRenderer.material.applyUniforms()
					}

				}

				nk.NkLayoutRowDynamic(ctxGUI, 25, 1)
				{
					nk.NkPropertyFloat(ctxGUI, "Rotation Speed: ", 0, &rotationSpeed, 10, 0.1, 0.1)
				}
				nk.NkLayoutRowDynamic(ctxGUI, 25, 1)
				{
					nk.NkPropertyFloat(ctxGUI, "Scale: ", 0.1, &scale, 2, 0.1, 0.1)
				}

				nk.NkLayoutRowDynamic(ctxGUI, 25, 2)
				{
					nk.NkLabel(ctxGUI, "Clearcolor: ", nk.TextLeft)
					size := nk.NkVec2(nk.NkWidgetWidth(ctxGUI), 400)
					if nk.NkComboBeginColor(ctxGUI, state.bgColor, size) > 0 {
						nk.NkLayoutRowDynamic(ctxGUI, 120, 1)
						state.bgColor = nk.NkColorPicker(ctxGUI, state.bgColor, nk.ColorFormatRGBA)
						nk.NkLayoutRowDynamic(ctxGUI, 25, 1)
						r, g, b, a := state.bgColor.RGBAi()
						r = nk.NkPropertyi(ctxGUI, "#R:", 0, r, 255, 1, 1)
						g = nk.NkPropertyi(ctxGUI, "#G:", 0, g, 255, 1, 1)
						b = nk.NkPropertyi(ctxGUI, "#B:", 0, b, 255, 1, 1)
						a = nk.NkPropertyi(ctxGUI, "#A:", 0, a, 255, 1, 1)
						clearColor = mgl32.Vec4{float32(r) / 255, float32(g) / 255, float32(b) / 255, float32(a) / 255}
						state.bgColor.SetRGBAi(r, g, b, a)
						nk.NkComboEnd(ctxGUI)
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
