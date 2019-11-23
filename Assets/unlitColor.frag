#version 330
struct Material {
    vec3 RGB;
};
uniform Material material;


out vec4 outputColor;
void main() {
    outputColor = vec4(material.RGB,1);
}