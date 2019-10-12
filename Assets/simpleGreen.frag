#version 330
in vec2 fragTexCoord;
out vec4 outputColor;
void main() {
    vec3 lightDir = vec3(1,-1,1);
    vec3 lightColor = vec3(1,1,1);
    
    outputColor = vec4(0,1,0,1);
}