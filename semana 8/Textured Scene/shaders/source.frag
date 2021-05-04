#version 410 core
in vec3 Normal;
in vec2 TexCoord;
out vec4 FragColor;
uniform vec3 objectColor;
uniform sampler2D texSampler3;
void main()
{
    FragColor = vec4(objectColor, 1.0); 
    // textured if texture is not empty, else colored
    if (textureSize(texSampler3, 0).x > 1){
            FragColor = texture(texSampler3, TexCoord) * vec4(objectColor, 1.0);
    }
    else {
        
        FragColor = vec4(objectColor, 1.0);
    }    
}