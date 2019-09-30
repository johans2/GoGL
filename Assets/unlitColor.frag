#version 330
out vec4 outputColor;
uniform vec3 RGB;
void main() {
    outputColor = vec4(RGB.r,RGB.g,RGB.b,1);
}