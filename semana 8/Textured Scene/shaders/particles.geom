#version 410 core
layout (points) in;
layout (triangle_strip, max_vertices = 800) out;

in VS_OUT {
    vec4 color;
} gs_in[];

in float seed[];

uniform mat4 projection;
uniform float particle_size;

out vec2 fUV;
out vec4 fColor;

/**
 * Generates random integer from a specified range.
 *
 * @param min    Minimal value
 * @param range  Range from minimal value
 *
 * @return Random integer in range min...min+range.
 */
int randomIntMinRange(int min, int range, float seed);


vec2[4] aFUV =  vec2[](vec2(0,0), vec2(0.5,0), vec2(0,0.5), vec2(0.5,0.5));
vec2[4] bFUV =  vec2[](vec2(0,0.5), vec2(1,0), vec2(0.5,0.5), vec2(1,0.5));
vec2[4] cFUV =  vec2[](vec2(0.5,0), vec2(0.5,0.5), vec2(0,1), vec2(0.5,1));
vec2[4] dFUV =  vec2[](vec2(0.5,0.5), vec2(1,0.5), vec2(0.5,1),vec2(1,1));


void main (void)
{
  vec4 P = gl_in[0].gl_Position;
  int index = randomIntMinRange(1,3, seed[0]);

  // a: left-bottom 
  vec2 va = P.xy + vec2(-0.5, -0.5) * particle_size;
  gl_Position = projection * vec4(va, P.zw);
  fUV = aFUV[index];
  fColor = gs_in[0].color;
  EmitVertex();  
  
  // b: left-top
  vec2 vb = P.xy + vec2(-0.5, 0.5) * particle_size;
  gl_Position = projection * vec4(vb, P.zw);
  fUV = bFUV[index];
  fColor = gs_in[0].color;
  EmitVertex();  
  
  // d: right-bottom
  vec2 vd = P.xy + vec2(0.5, -0.5) * particle_size;
  gl_Position = projection * vec4(vd, P.zw);
  fUV = cFUV[index];
  fColor = gs_in[0].color;
  EmitVertex();  

  // c: right-top
  vec2 vc = P.xy + vec2(0.5, 0.5) * particle_size;
  gl_Position = projection * vec4(vc, P.zw);
  fUV = dFUV[index];
  fColor = gs_in[0].color;
  EmitVertex();  

  EndPrimitive();  
}


int randomIntMinRange(int min, int range, float seed)
{
    vec3 currentRandomGeneratorSeed = vec3(seed,0,0);
    uint n = floatBitsToUint(currentRandomGeneratorSeed.y * 214013.0 + currentRandomGeneratorSeed.x * 2531011.0 + currentRandomGeneratorSeed.z * 141251.0);
    n = n * (n * n * 15731u + 789221u);
    n = (n >> 9u) | 0x3F800000u;
    float result =  2.0 - uintBitsToFloat(n);

    currentRandomGeneratorSeed = vec3(currentRandomGeneratorSeed.x + 147158.0 * result,
        currentRandomGeneratorSeed.y * result  + 415161.0 * result,
        currentRandomGeneratorSeed.z + 324154.0 * result);

    return min + int(n) % range;
}