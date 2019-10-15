#version 330
uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projMatrix;
uniform mat4 MVP;

in vec3 vert;
in vec2 vertTexCoord;
in vec3 normal;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vertTexCoord;
	gl_Position = projMatrix * viewMatrix * modelMatrix * vec4(vert, 1);
}