#version 330
in vec2 fragTexCoord;
in vec3 fragNormal;
in vec3 fragVert;
in vec3 worldPos;

uniform mat4 modelMatrix;
uniform vec3 cameraWorldPos;
out vec4 outputColor;

void main() {
    //calculate normal in world coordinates
    mat3 worldMatrix = transpose(inverse(mat3(modelMatrix)));
    vec3 normal = normalize(worldMatrix * fragNormal);

    vec4 color = vec4(0.2,0.2,0.9,1);
    vec4 indirectDiffuse = vec4(0.2,0.2,0.2,1);

    vec3 lightColor = vec3(1,1,1) * 0.6;
    vec4 lightDir = vec4(0.5,1.2,1.5,1);
    vec3 directDiffuse = lightColor * dot(normal, normalize(lightDir.xyz));
    vec4 diffuse = indirectDiffuse + vec4(directDiffuse,1);
    
    outputColor = color * diffuse;// + vec4(lightColor,1) * specular;
}