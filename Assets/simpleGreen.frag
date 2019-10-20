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

    vec4 color = vec4(0.2,0.9,0.2,1);
    vec4 indirectDiffuse = vec4(0.65,0.65,0.65,1);

    vec3 lightColor = vec3(1,1,1) * 0.6;
    vec4 lightDir = vec4(-0.5,0.1,-1,1);
    vec3 directDiffuse = lightColor * dot(normal, normalize(lightDir.xyz));
    vec4 diffuse = indirectDiffuse + vec4(directDiffuse,1);

    outputColor = color * diffuse;
}