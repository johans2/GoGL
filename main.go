package main

import (
	"fmt"
	"go/build"
	_ "image/png"
	"log"
	"os"
	"runtime"

	"GoGL/gui"
	"GoGL/platform"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go"
)

const windowWidth = 1600
const windowHeight = 900

type mouseState bool

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

type state struct {
	rotationSpeed  float32
	scale          float32
	clearColorR    float32
	clearColorG    float32
	clearColorB    float32
	vertSource     string
	fragSource     string
	activeMaterial material
	activeModel    []float32
	modelRenderer  renderer
	shaderError    error
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
	context, imguiInput := gui.NewImgui()
	defer context.Destroy()

	// Setup the GLFW platform
	platform, err := platform.NewPlatform(windowWidth, windowHeight)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer platform.Dispose()

	// Setup the Imgui renderer
	imguiRenderer, err := gui.NewOpenGL3()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer imguiRenderer.Dispose()

	// Setup the platform callbacks
	platform.SetMouseButtonCallback(imguiInput.MouseButtonChange)
	platform.SetScrollCallback(imguiInput.MouseScrollChange)
	platform.SetKeyCallback(imguiInput.KeyChange)
	platform.SetCharCallback(imguiInput.CharChange)

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Create state and data holders
	state := new(state)
	data := new(data)

	var defaultShader shader
	defaultShader.loadFromFile("Assets/simpleGreen.vert", "Assets/simpleGreen.frag")

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
	state.activeMaterial.init(defaultShader)
	state.activeModel = data.boxVerts
	state.modelRenderer.setData(state.activeModel, state.activeMaterial)
	state.vertSource = ""
	state.fragSource = ""
	state.clearColorR = 1
	state.clearColorG = 1
	state.clearColorB = 1
	state.rotationSpeed = float32(0.5)
	state.scale = float32(1.0)

	for !platform.ShouldStop() {
		platform.ProcessEvents()
		cursorX, cursorY := platform.GetCursorPos()

		mouseState := gui.ImguiMouseState{
			MousePosX:  float32(cursorX),
			MousePosY:  float32(cursorY),
			MousePress: platform.GetMousePresses123()}

		imguiInput.NewFrame(platform.DisplaySize()[0], platform.DisplaySize()[1], glfw.GetTime(), platform.IsFocused(), mouseState)
		{
			imgui.Begin("Material viewer")

			imgui.Text("Shader programs")
			imgui.Text("vert source")
			imgui.SameLine()
			imgui.InputText("##vert source", &state.vertSource)
			imgui.Text("frag source")
			imgui.SameLine()
			imgui.InputText("##frag source", &state.fragSource)

			imgui.Columns(3, "")
			imgui.NextColumn()
			if imgui.ButtonV("Compile", imgui.Vec2{X: 100, Y: 30}) {
				var newShader shader
				state.shaderError = newShader.loadFromFile(state.vertSource, state.fragSource)
				if state.shaderError == nil {
					var newMaterial material
					newMaterial.init(newShader)
					state.activeMaterial = newMaterial
					state.modelRenderer.setData(state.activeModel, state.activeMaterial)
				} else {
					log.Printf("ERROR: " + (state.shaderError).Error())
				}
			}
			imgui.Columns(1, "")

			if state.shaderError != nil {
				//imgui.Text(state.shaderError.Error())
				err := state.shaderError.Error()
				imgui.InputTextMultiline("##shaderError", &err)
			}

			// Draw the material GUI
			if len(state.activeMaterial.fields) != 0 || len(state.activeMaterial.texBindings) != 0 {

				state.modelRenderer.material.drawUI()

				if imgui.Button("Apply") {
					state.modelRenderer.material.applyUniforms()
				}
			}

			imgui.End()
			imgui.Begin("Global properties")
			drawUtilityGUI(state, data)
			imgui.End()
		}

		// Rendering
		imgui.Render() // This call only creates the draw data list. Actual rendering to framebuffer is done below.

		// Need to reanable these things since Nuklear sets its own gl states when rendering.
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.CULL_FACE)
		gl.DepthFunc(gl.LESS)
		gl.ClearColor(state.clearColorR,
			state.clearColorG,
			state.clearColorB,
			1)

		// Update time
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += (elapsed * float64(state.rotationSpeed))

		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		model = model.Mul4(mgl32.Scale3D(state.scale, state.scale, state.scale))

		// Set global rendering properties
		GlobalRenderProps.CameraPos = [3]float32{cameraPos.X(), cameraPos.Y(), cameraPos.Z()}
		GlobalRenderProps.Time = float32(time)
		ApplyGlobalRenderProperties(state.activeMaterial.shader.program)

		// Render the model
		state.modelRenderer.issueDrawCall(model, view, projection, cameraPos, float32(time))

		// Maintenance
		imguiRenderer.Render(platform.DisplaySize(), platform.FramebufferSize(), imgui.RenderedDrawData())
		platform.PostRender()
	}
}

func drawShaderInputGUI(state *state) {
	imgui.Text("Shader programs")
	imgui.Text("vert source")
	imgui.SameLine()
	imgui.InputText("##vert source", &state.vertSource)
	imgui.Text("frag source")
	imgui.SameLine()
	imgui.InputText("##frag source", &state.fragSource)

	imgui.Columns(3, "")
	imgui.NextColumn()
	if imgui.ButtonV("Compile", imgui.Vec2{X: 100, Y: 30}) {
		var newShader shader
		state.shaderError = newShader.loadFromFile(state.vertSource, state.fragSource)
		if state.shaderError == nil {
			var newMaterial material
			newMaterial.init(newShader)
			state.activeMaterial = newMaterial
			state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		} else {
			log.Printf("ERROR: " + (state.shaderError).Error())
		}
	}
	imgui.Columns(1, "")

	if state.shaderError != nil {
		//imgui.Text(state.shaderError.Error())
		err := state.shaderError.Error()
		imgui.InputTextMultiline("##shaderError", &err)
	}

}

// Draw the utility functions GUI.
func drawUtilityGUI(state *state, data *data) {
	imgui.Columns(4, "")
	if imgui.Button("	Sphere	") {
		state.activeModel = data.sphereVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
	}
	imgui.NextColumn()
	if imgui.Button("	Box		") {
		state.activeModel = data.boxVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
	}
	imgui.NextColumn()
	if imgui.Button("	Torus	") {
		state.activeModel = data.torusVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
	}
	imgui.NextColumn()
	if imgui.Button("	Plane	") {
		state.activeModel = data.planeVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
	}
	imgui.Columns(1, "")

	imgui.Columns(4, "")
	imgui.Text("Clear color:")
	imgui.NextColumn()
	imgui.SliderFloat("R", &state.clearColorR, 0, 1)
	imgui.NextColumn()
	imgui.SliderFloat("G", &state.clearColorG, 0, 1)
	imgui.NextColumn()
	imgui.SliderFloat("B", &state.clearColorB, 0, 1)
	imgui.Columns(1, "")
	imgui.Text("Rotation speed:")
	imgui.SameLine()
	imgui.SliderFloat("##rotSpeed", &state.rotationSpeed, 0, 10)

	imgui.Text("LightDir")
	imgui.SameLine()
	if imgui.SliderFloat3("##lightDir", &GlobalRenderProps.LightDir, -365, 365) {
		ApplyLightColor(state.activeMaterial.shader.program)
	}

	imgui.Text("LightColor")
	imgui.SameLine()
	if imgui.SliderFloat3("##lightCol", &GlobalRenderProps.LightColor, 0, 1) {
		ApplyLightColor(state.activeMaterial.shader.program)
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
