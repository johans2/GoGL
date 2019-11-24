package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-gl/gl/v3.2-core/gl"
)

type uniformType string

const (
	uniformFloat uniformType = "float"
	uniformVec2  uniformType = "vec2"
	uniformVec3  uniformType = "vec3"
	uniformVec4  uniformType = "vec4"
	uniformTex2D uniformType = "sampler2D"
	uniformMat4  uniformType = "mat4"
)

const (
	camWorldPosName string = "cameraWorldPos"
	modelMatrixName string = "modelMatrix"
	viewMatrixName  string = "viewMatrix"
	projMatrixName  string = "projMatrix"
	mvpMatrixName   string = "MVP"
	timeName        string = "time"
)

type shader struct {
	program    uint32
	vertSource string
	fragSource string
	uniforms   []uniform
}

type uniform struct {
	uType uniformType
	name  string
}

func getUniforms(source string) []uniform {
	uniforms := make([]uniform, 0)

	lines := strings.Split(source, "\n")
	startMaterialStruct := false
	for _, line := range lines {
		if !startMaterialStruct && (strings.Contains(line, "struct Material") || strings.Contains(line, "struct material")) {
			startMaterialStruct = true
			continue
		}

		if startMaterialStruct {
			words := strings.Split(strings.Trim(line, " "), " ")
			if strings.Contains(line, "};") {
				break
			}

			uType, error := getUniformTypeFromString(strings.TrimSpace(words[0]))
			if error != nil {
				fmt.Println(error.Error())
			}

			name := strings.TrimSpace(strings.Replace(words[1], ";", "", -1))
			u := uniform{uType, name}
			uniforms = append(uniforms, u)
		}
	}
	return uniforms
}

func isReservedUniformName(word string) bool {
	isReserved := word == camWorldPosName ||
		word == modelMatrixName ||
		word == viewMatrixName ||
		word == projMatrixName ||
		word == mvpMatrixName

	return isReserved
}

// GetUniformTypeFromString Get the uniform type form a shader word
func getUniformTypeFromString(word string) (uniformType, error) {
	switch word {
	case "float":
		return uniformFloat, nil
	case "vec2":
		return uniformVec2, nil
	case "vec3":
		return uniformVec3, nil
	case "vec4":
		return uniformVec4, nil
	case "sampler2D":
		return uniformTex2D, nil
	case "mat4":
		return uniformMat4, nil
	default:
		return "", fmt.Errorf("Unsupported shader uniform: %s", word)
	}
}

func (s *shader) loadFromFile(vertSource string, fragSource string) error {
	vertFile, errV := os.Open(vertSource)
	fragFile, errF := os.Open(fragSource)
	defer vertFile.Close()
	defer fragFile.Close()

	if errV != nil {
		return errV
	}
	if errF != nil {
		return errF
	}

	vertBytes, errV := ioutil.ReadAll(vertFile)
	fragBytes, errF := ioutil.ReadAll(fragFile)

	s.vertSource = string(vertBytes) + "\x00"
	s.fragSource = string(fragBytes) + "\x00"

	var compileErr error
	s.program, compileErr = newProgram(s.vertSource, s.fragSource)

	if compileErr != nil {
		return compileErr
	}

	s.uniforms = append(s.uniforms, getUniforms(s.vertSource)...)
	s.uniforms = append(s.uniforms, getUniforms(s.fragSource)...)
	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}
