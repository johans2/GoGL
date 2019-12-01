#version 330
struct Material {
    float animStr;
    sampler2D noise;
};

uniform Material material;

// Universal uniforms
uniform float time;
uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projMatrix;
uniform mat4 MVP;

in vec3 vert;
in vec2 vertTexCoord;
in vec3 normal;
out vec2 fragTexCoord;
out vec3 fragNormal;
out vec3 fragVert;
out vec3 fragWorldPos;
void main() {
    fragTexCoord = vertTexCoord;
    fragNormal = normal;
    fragVert = vert;

    fragTexCoord.x += time;

    vec4 noiseVal = texture(material.noise, fragTexCoord);
    fragVert += fragNormal * noiseVal.x * material.animStr;

    fragWorldPos = (modelMatrix * vec4(fragVert,1)).xyz;
	gl_Position = MVP * vec4(fragVert, 1);
}