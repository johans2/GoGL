#version 330
struct Material {
    sampler2D tex;
};

uniform Material material;

in vec2 fragTexCoord;
out vec4 outputColor;
void main() {
    vec4 texColor = texture(material.tex, fragTexCoord);
    outputColor = texColor;
}