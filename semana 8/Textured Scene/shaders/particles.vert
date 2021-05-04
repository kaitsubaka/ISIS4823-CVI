#version 410 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec4 aColor;

out VS_OUT {
    vec4 color;
} vs_out;

out float seed;

uniform mat4 model;
uniform mat4 view;

void main()
{
    vs_out.color = aColor;
    seed = aPos.x;
    gl_Position = model * view * vec4(aPos, 1.0);
}