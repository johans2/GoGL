#version 330
struct Material {
    vec3 a_wl_s_1;
    vec3 a_wl_s_2;
    vec2 dir1;
    vec2 dir2;
    float steepness1;
    float steepness2;
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
out vec3 worldPos;


// Returns x,y,z position and w crestFactor (used for foam)
vec4 WavePoint(vec2 position, float amplitude, float wavelength, float speed, vec2 direction, float steepness) {
    float frequency = 2 / wavelength;
    float phaseConstantSpeed = speed * 2 / wavelength;
	

	vec2 normalizedDir = normalize(direction);
    float fi = time  * phaseConstantSpeed;
    float dirDotPos = dot(normalizedDir, position);

    float waveGretsX = steepness * amplitude * normalizedDir.x * cos(frequency * dirDotPos + fi);
	float crest = sin(frequency * dirDotPos + fi);
    float waveGretsY = amplitude * crest;
    float waveGretsZ = steepness * amplitude * normalizedDir.y * cos(frequency * dirDotPos + fi);
	float crestFactor = crest * clamp(steepness,0.0,1.0);

    return vec4(waveGretsX, waveGretsY, waveGretsZ, crestFactor);
}

vec3 WaveNormal(vec3 position, float amplitude, float wavelength, float speed, vec2 direction, float steepness) {

	float frequency = 2 / wavelength;
	float phaseConstantSpeed = speed * 2 / wavelength;

	vec2 normalizedDir = normalize(direction);
	float fi = time  * phaseConstantSpeed;
	float dirDotPos = dot(normalizedDir, position.xz);

	float WA = frequency * amplitude;
	float S = sin(frequency * dirDotPos + fi);
	float C = cos(frequency * dirDotPos + fi);

	vec3 normal = vec3 (
		normalizedDir.x * WA * C,
		min(0.2f,steepness * WA * S),
		normalizedDir.y * WA * C
	);

	return normal;
}

void main() {
    fragTexCoord = vertTexCoord;
    fragNormal = normal;
    fragVert = vert;
    worldPos = (modelMatrix * vec4(vert,1)).xyz;
    worldPos += WavePoint(worldPos.xz, 
                            material.a_wl_s_1.x, 
                            material.a_wl_s_1.y, 
                            material.a_wl_s_1.z, 
                            material.dir1, 
                            material.steepness1).xyz;

    worldPos += WavePoint(worldPos.xz, 
                            material.a_wl_s_2.x, 
                            material.a_wl_s_2.y, 
                            material.a_wl_s_2.z, 
                            material.dir2, 
                            material.steepness2).xyz;

    fragNormal = WaveNormal(worldPos.xyz, 
                            material.a_wl_s_1.x, 
                            material.a_wl_s_1.y, 
                            material.a_wl_s_1.z, 
                            material.dir1, 
                            material.steepness1);

    fragNormal += WaveNormal(worldPos.xyz, 
                            material.a_wl_s_2.x, 
                            material.a_wl_s_2.y, 
                            material.a_wl_s_2.z, 
                            material.dir2, 
                            material.steepness2);

	gl_Position = MVP * vec4(vert + worldPos.xyz, 1);
}
