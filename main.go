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
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
)

const windowWidth = 1600
const windowHeight = 900

type mouseState bool

const (
	press   mouseState = true
	release mouseState = false
)

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

type state struct {
	glContext        *nk.Context
	bgColor          nk.Color
	rotationSpeed    float32
	scale            float32
	clearColor       mgl32.Vec4
	bufferVertSource []byte
	bufferFragSource []byte
	activeMaterial   material
	activeModel      []float32
	modelRenderer    renderer
	shaderError      error
}

type data struct {
	sphereVerts []float32
	boxVerts    []float32
	torusVerts  []float32
	planeVerts  []float32
	coneVerts   []float32
}

var reApplyUniformsa = false

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()

	// Set the working directory to the root of Go package, so that its assets can be accessed.
	dir, err := importPathToDir("GoGL")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

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

	// Create state and data holders
	state := new(state)
	data := new(data)

	// Init nuklear gui
	log.Printf("NkPlatformInit")
	state.glContext = nk.NkPlatformInit(window, nk.PlatformInstallCallbacks)

	// Create font
	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(state.glContext, sansFont.Handle())
	}
	log.Println("Finished setting up Nk-GUI")

	var shaderGreen shader
	shaderGreen.loadFromFile("Assets/simpleGreen.vert", "Assets/simpleGreen.frag")

	// Set up projection matrix for shader
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10.0)

	// Set up view matrix for shader
	cameraPos := mgl32.Vec3{0, 2, 3}
	view := mgl32.LookAtV(cameraPos, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	// Set up model martix for shader
	model := mgl32.Ident4()

	// Load the model from the obj file
	sphereModel, _ := readOBJ("Assets/sphere.obj")
	boxModel, _ := readOBJ("Assets/box.obj")
	torusModel, _ := readOBJ("Assets/torus.obj")
	planeModel, _ := readOBJ("Assets/plane.obj")
	coneModel, _ := readOBJ("Assets/cone.obj")

	angle := 0.0
	previousTime := glfw.GetTime()

	log.Printf("Finished setup. Now rendering..")

	// Setup data struct
	data.sphereVerts = sphereModel.ToArrayXYZUVN1N2N3()
	data.boxVerts = boxModel.ToArrayXYZUVN1N2N3()
	data.torusVerts = torusModel.ToArrayXYZUVN1N2N3()
	data.planeVerts = planeModel.ToArrayXYZUVN1N2N3()
	data.coneVerts = coneModel.ToArrayXYZUVN1N2N3()

	// Setup initial state
	state.activeMaterial.init(shaderGreen)
	state.activeModel = data.boxVerts
	state.modelRenderer.setData(state.activeModel, state.activeMaterial)
	state.bufferVertSource = make([]byte, 1024)
	state.bufferFragSource = make([]byte, 1024)
	state.bgColor.SetRGBAi(255, 255, 255, 255)
	state.clearColor = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	state.rotationSpeed = float32(0.5)
	state.scale = float32(1.0)

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
		gl.ClearColor(state.clearColor.X(),
			state.clearColor.Y(),
			state.clearColor.Z(),
			state.clearColor.W())

		// Update time
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		// Update mouse rotation
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
			angle += (elapsed * float64(state.rotationSpeed))
		}

		if reApplyUniformsa {
			state.modelRenderer.material.applyUniforms()
		}

		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		model = model.Mul4(mgl32.Scale3D(state.scale, state.scale, state.scale))

		// Render
		state.modelRenderer.issueDrawCall(model, view, projection, cameraPos, float32(time))
		drawGUI(state, data, window)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}

	nk.NkPlatformShutdown()
	glfw.Terminate()
}

func drawGUI(state *state, data *data, window *glfw.Window) {
	nk.NkPlatformNewFrame()
	bounds := nk.NkRect(20, 20, 400, 800)
	update := nk.NkBegin(state.glContext, "Material inspector", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

	if update > 0 {
		nk.NkLayoutRowStatic(state.glContext, 10, 80, 1)
		{
			// Draw shader source and compilation GUI
			nk.NkLayoutRowDynamic(state.glContext, 30, 1)
			{
				nk.NkLabel(state.glContext, "SHADER FILES", nk.TextCentered)
				nk.NkLabel(state.glContext, "Vertex program:", nk.TextLeft)
				nk.NkEditStringZeroTerminated(state.glContext, nk.EditField, state.bufferVertSource, 256, nk.NkFilterDefault)
				nk.NkLabel(state.glContext, "Fragment program:", nk.TextLeft)
				nk.NkEditStringZeroTerminated(state.glContext, nk.EditField, state.bufferFragSource, 256, nk.NkFilterDefault)
				if nk.NkButtonLabel(state.glContext, "Compile") > 0 {
					var newShader shader
					nVert := bytes.IndexByte(state.bufferVertSource, 0)
					pathStringVert := string(state.bufferVertSource[:nVert])
					nFrag := bytes.IndexByte(state.bufferFragSource, 0)
					pathStringFrag := string(state.bufferFragSource[:nFrag])

					state.shaderError = newShader.loadFromFile(pathStringVert, pathStringFrag)
					if state.shaderError == nil {
						var newMaterial material
						newMaterial.init(newShader)
						state.activeMaterial = newMaterial
						state.modelRenderer.setData(state.activeModel, state.activeMaterial)
					} else {
						log.Printf("ERROR: " + (state.shaderError).Error())
					}
				}

				if (*state).shaderError != nil {
					nk.NkLayoutRowDynamic(state.glContext, 60, 1)
					{
						nk.NkLabelWrap(state.glContext, "ERROR: "+state.shaderError.Error())
					}
				}
			}
			// Draw the material GUI
			if len(state.activeMaterial.fields) != 0 || len(state.activeMaterial.texBindings) != 0 {
				nk.NkLayoutRowDynamic(state.glContext, 30, 1)
				{
					nk.NkLabel(state.glContext, "SHADER PROPERTIES", nk.TextCentered)
				}

				state.modelRenderer.material.drawUI(state.glContext)

				nk.NkLayoutRowDynamic(state.glContext, 30, 1)
				{
					if nk.NkButtonLabel(state.glContext, "Apply") > 0 {
						reApplyUniformsa = true
					}
				}
			}

			// Draw the model picker GUI
			nk.NkLabel(state.glContext, "", nk.TextCentered)
			nk.NkLabel(state.glContext, "MODEL", nk.TextCentered)
			nk.NkLayoutRowDynamic(state.glContext, 60, 5)
			{
				if nk.NkButtonLabel(state.glContext, "Sphere") > 0 {
					state.activeModel = data.sphereVerts
					state.modelRenderer.setData(state.activeModel, state.activeMaterial)
					state.modelRenderer.material.applyUniforms()
				}
				if nk.NkButtonLabel(state.glContext, "Box") > 0 {
					state.activeModel = data.boxVerts
					state.modelRenderer.setData(state.activeModel, state.activeMaterial)
					state.modelRenderer.material.applyUniforms()
				}
				if nk.NkButtonLabel(state.glContext, "Torus") > 0 {
					state.activeModel = data.torusVerts
					state.modelRenderer.setData(state.activeModel, state.activeMaterial)
					state.modelRenderer.material.applyUniforms()
				}
				if nk.NkButtonLabel(state.glContext, "Plane") > 0 {
					state.activeModel = data.planeVerts
					state.modelRenderer.setData(state.activeModel, state.activeMaterial)
					state.modelRenderer.material.applyUniforms()
				}
				if nk.NkButtonLabel(state.glContext, "Cone") > 0 {
					state.activeModel = data.coneVerts
					state.modelRenderer.setData(state.activeModel, state.activeMaterial)
					state.modelRenderer.material.applyUniforms()
				}

			}

			// Draw utility values GUI
			nk.NkLayoutRowDynamic(state.glContext, 25, 1)
			{
				nk.NkPropertyFloat(state.glContext, "Rotation Speed: ", 0, &state.rotationSpeed, 10, 0.1, 0.1)
			}
			nk.NkLayoutRowDynamic(state.glContext, 25, 1)
			{
				nk.NkPropertyFloat(state.glContext, "Scale: ", 0.1, &state.scale, 2, 0.1, 0.1)
			}

			nk.NkLayoutRowDynamic(state.glContext, 25, 2)
			{
				nk.NkLabel(state.glContext, "Clearcolor: ", nk.TextLeft)
				size := nk.NkVec2(nk.NkWidgetWidth(state.glContext), 400)
				if nk.NkComboBeginColor(state.glContext, state.bgColor, size) > 0 {
					nk.NkLayoutRowDynamic(state.glContext, 120, 1)
					state.bgColor = nk.NkColorPicker(state.glContext, state.bgColor, nk.ColorFormatRGBA)
					nk.NkLayoutRowDynamic(state.glContext, 25, 1)
					r, g, b, a := state.bgColor.RGBAi()
					r = nk.NkPropertyi(state.glContext, "#R:", 0, r, 255, 1, 1)
					g = nk.NkPropertyi(state.glContext, "#G:", 0, g, 255, 1, 1)
					b = nk.NkPropertyi(state.glContext, "#B:", 0, b, 255, 1, 1)
					a = nk.NkPropertyi(state.glContext, "#A:", 0, a, 255, 1, 1)
					state.clearColor = mgl32.Vec4{float32(r) / 255, float32(g) / 255, float32(b) / 255, float32(a) / 255}
					state.bgColor.SetRGBAi(r, g, b, a)
					nk.NkComboEnd(state.glContext)
				}

			}

		}

	}
	nk.NkEnd(state.glContext)

	// Render GUI
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height := window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
}

func initGLFW() *glfw.Window {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "GoGL", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
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
