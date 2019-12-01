#version 330
in vec2 fragTexCoord;
in vec3 fragNormal;
in vec3 fragVert;
in vec3 fragWorldPos;

uniform mat4 modelMatrix;
uniform vec3 cameraWorldPos;

out vec4 outputColor;
void main() {
    outputColor = vec4(0.1,0.9,0.3,1);
}