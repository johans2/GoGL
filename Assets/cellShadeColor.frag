#version 330
struct Material {
    vec3 color;
    float specPower;
    sampler2D cellRampDiffuse;
    sampler2D cellRampSpecular;
}; 
  
uniform Material material;

in vec2 fragTexCoord;
in vec3 fragNormal;
in vec3 fragWorldPos;

uniform mat4 modelMatrix;
uniform vec3 cameraWorldPos;

out vec4 outputColor;
void main() {
    // Calculate normal in world coordinates
    mat3 worldMatrix = transpose(inverse(mat3(modelMatrix)));
    vec3 normal = normalize(worldMatrix * fragNormal);

    // Calculate diffuse light
    vec4 indirectDiffuse = vec4(0.2,0.2,0.2,1);
    vec3 lightColor = vec3(1,1,1) * 0.6;
    vec4 lightDir = vec4(0.5,1.2,1.5,1);
    float nDotL = dot(normal, normalize(lightDir.xyz));
    float cellLight = clamp(texture(material.cellRampDiffuse, vec2(nDotL,0)).r,0,1);
    vec3 directDiffuse = lightColor * cellLight;    
    vec4 diffuse = indirectDiffuse + vec4(directDiffuse,1);

    // Calculate specular highlight
    vec3 viewDir = normalize(fragWorldPos - cameraWorldPos);
    vec3 halfDir = normalize(lightDir.xyz + viewDir);
    float specAngle = max(dot(halfDir, normal), 0.0);
    float specular = pow(specAngle,material.specPower);
    float cellSpecular = clamp(texture(material.cellRampSpecular, vec2(specular,0)).r,0,1);

    outputColor = vec4(material.color,1) * diffuse + vec4(lightColor,1) * cellSpecular;
}