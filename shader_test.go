package main

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetShaderUniforms(t *testing.T) {

	var expectedUniforms []uniform
	testUniformFloat := uniform{uniformFloat, "testFloat"}
	testUniformVec2 := uniform{uniformVec2, "testVec2"}
	testUniformVec3 := uniform{uniformVec3, "testVec3"}
	testUniformVec4 := uniform{uniformVec4, "testVec4"}
	testUniformTex2D := uniform{uniformTex2D, "testTex2D"}

	expectedUniforms = append(expectedUniforms,
		testUniformFloat,
		testUniformVec2,
		testUniformVec3,
		testUniformVec4,
		testUniformTex2D)

	testShader := `#version 330
	struct Material {
		float testFloat;
		vec2 testVec2;
		vec3 testVec3;
		vec4 testVec4;
		sampler2D testTex2D;
	}; 
	
	in vec3 vert;
	in vec2 vertTexCoord;
	in vec3 normal;
	out vec2 fragTexCoord;
	void main() {
		fragTexCoord = vertTexCoord;
		gl_Position = projMatrix * viewMatrix * modelMatrix * vec4(vert, 1);
	}`

	shaderUniforms := getUniforms(testShader)
	assert.Equal(t, len(shaderUniforms), len(expectedUniforms), "Invalid number of shader material fields parsed.")

	for i, _ := range shaderUniforms {
		expeced := expectedUniforms[i]
		actual := shaderUniforms[i]
		assert.Equal(t, expeced.name, actual.name, "Invalid name on shader material field")
		assert.Equal(t, expeced.uType, actual.uType, "Invalid type on shader material field")
	}

}
