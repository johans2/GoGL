#version 330
struct Material {
    vec3 color;
}; 
  
uniform Material material;

out vec4 outputColor;
void main() {
    outputColor = vec4(material.color,1);
}