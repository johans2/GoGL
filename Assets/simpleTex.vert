#version 330
uniform mat4 MVP;
in vec3 vert;
in vec2 vertTexCoord;
in vec3 normal;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vertTexCoord;
	gl_Position = MVP * vec4(vert, 1);
}
