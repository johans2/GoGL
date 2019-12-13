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
	press   mouseState = true
	release mouseState = false
)

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

type state struct {
	rotationSpeed  float32
	scale          float32
	clearColor     mgl32.Vec4
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
	platform, err := platform.NewPlatform()
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

	currentMouseState := release
	// Initialize Glow
	/*if err := gl.Init(); err != nil {
		panic(err)
	}*/

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Create state and data holders
	state := new(state)
	data := new(data)

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
	state.vertSource = ""
	state.fragSource = ""
	state.clearColor = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	state.rotationSpeed = float32(0.5)
	state.scale = float32(1.0)

	var mouseXPrev float64
	var mouseYPrev float64
	var mouseY float64
	var mouseX float64

	for !platform.ShouldStop() {
		platform.ProcessEvents()
		cursorX, cursorY := platform.GetCursorPos()
		// Signal start of a new frame
		//platform.NewFrame()
		//imgui.NewFrame()

		mouseState := gui.ImguiMouseState{
			float32(cursorX),
			float32(cursorY),
			[3]bool{platform.GetMousePress(glfw.MouseButton1),
				platform.GetMousePress(glfw.MouseButton2),
				platform.GetMousePress(glfw.MouseButton3)}}

		imguiInput.NewFrame(platform.DisplaySize()[0], platform.DisplaySize()[1], glfw.GetTime(), platform.IsFocused(), mouseState)
		{
			imgui.Begin("Material viewer")

			imgui.Text("Shader programs")
			imgui.InputText("vert source", &state.vertSource)
			imgui.InputText("frag source", &state.fragSource)

			if imgui.Button("Compile") {
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

			// Draw the material GUI
			if len(state.activeMaterial.fields) != 0 || len(state.activeMaterial.texBindings) != 0 {

				state.modelRenderer.material.drawUI()

				if imgui.Button("Apply") {
					state.modelRenderer.material.applyUniforms()
				}
			}

			drawModelGUI(state, data)
			imgui.End()
		}

		// Rendering
		imgui.Render() // This call only creates the draw data list. Actual rendering to framebuffer is done below.

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
			x, y := 0.0, 0.0 //window.GetCursorPos()
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

		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		model = model.Mul4(mgl32.Scale3D(state.scale, state.scale, state.scale))

		// Render the model
		state.modelRenderer.issueDrawCall(model, view, projection, cameraPos, float32(time))

		// Maintenance
		imguiRenderer.Render(platform.DisplaySize(), platform.FramebufferSize(), imgui.RenderedDrawData())
		platform.PostRender()
	}
}

func drawModelGUI(state *state, data *data) {

	if imgui.Button("Sphere") {
		state.activeModel = data.sphereVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
	}
	imgui.SameLine()
	if imgui.Button("Box") {
		state.activeModel = data.boxVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
	}
	imgui.SameLine()
	if imgui.Button("Torus") {
		state.activeModel = data.torusVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
	}
	imgui.SameLine()
	if imgui.Button("Plane") {
		state.activeModel = data.planeVerts
		state.modelRenderer.setData(state.activeModel, state.activeMaterial)
		state.modelRenderer.material.applyUniforms()
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
