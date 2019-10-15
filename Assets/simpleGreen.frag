#version 330
in vec2 fragTexCoord;
in vec3 fragNormal;
in vec3 fragVert;

uniform mat4 modelMatrix;
out vec4 outputColor;
void main() {
    //calculate normal in world coordinates
    mat3 normalMatrix = transpose(inverse(mat3(modelMatrix)));
    vec3 normal = normalize(normalMatrix * fragNormal);


    vec3 lightColor = vec3(1,1,1) * 0.4;
    vec4 lightDir = vec4(1,0,0,1);
    vec3 light = lightColor * dot(normal, normalize(lightDir.xyz));

    outputColor = vec4(0.2,0.9,0.2,1) + vec4(light, 1);
}