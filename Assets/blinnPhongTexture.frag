#version 330
struct Material {
    float specPower;
    sampler2D tex;
}; 
  
uniform Material material;

in vec2 fragTexCoord;
in vec3 fragNormal;
in vec3 fragVert;
in vec3 fragWorldPos;

uniform mat4 modelMatrix;
uniform vec3 cameraWorldPos;

out vec4 outputColor;
void main() {
    // Calculate normal in world coordinates
    mat3 worldMatrix = transpose(inverse(mat3(modelMatrix)));
    vec3 normal = normalize(worldMatrix * fragNormal);

    // Sample texture for color
    vec4 color = texture(material.tex, fragTexCoord);

    // Calculate diffuse light
    vec4 indirectDiffuse = vec4(0.2,0.2,0.2,1);
    vec3 lightColor = vec3(1,1,1) * 0.6;
    vec4 lightDir = vec4(0.5,1.2,1.5,1);
    vec3 directDiffuse = lightColor * dot(normal, normalize(lightDir.xyz));
    vec4 diffuse = indirectDiffuse + vec4(directDiffuse,1);

    // Calculate specular highlight
    float specPower = 50;
    vec3 viewDir = normalize(fragWorldPos - cameraWorldPos);
    vec3 halfDir = normalize(lightDir.xyz + viewDir);
    float specAngle = max(dot(halfDir, normal), 0.0);
    float specular = pow(specAngle,material.specPower);

    outputColor = color * diffuse + vec4(lightColor,1) * specular;
}