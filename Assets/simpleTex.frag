#version 330
uniform sampler2D tex;
uniform sampler2D maskTex;
in vec2 fragTexCoord;
out vec4 outputColor;
void main() {
    vec4 texColor = texture(tex, fragTexCoord);
    vec4 maskColor = texture(maskTex, fragTexCoord);

    outputColor = texture(tex, fragTexCoord) * maskColor;
}