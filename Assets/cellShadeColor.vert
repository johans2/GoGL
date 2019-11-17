#version 330
uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projMatrix;
uniform mat4 MVP;

in vec3 vert;
in vec2 vertTexCoord;
in vec3 normal;
out vec2 fragTexCoord;
out vec3 fragNormal;
out vec3 fragWorldPos;

void main() {
    fragTexCoord = vertTexCoord;
    fragNormal = normal;
    fragWorldPos = (modelMatrix * vec4(vert,1)).xyz;
	gl_Position = MVP * vec4(vert, 1);
}