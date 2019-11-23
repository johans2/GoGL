#version 330
struct Material {
    uniform sampler2D tex;
    uniform sampler2D maskTex;
};

uniform Material material;

in vec2 fragTexCoord;
out vec4 outputColor;
void main() {
    vec4 texColor = texture(material.tex, fragTexCoord);
    vec4 maskColor = texture(material.maskTex, fragTexCoord);

    outputColor = texColor * maskColor;
}