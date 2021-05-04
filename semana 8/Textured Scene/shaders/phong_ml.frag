#version 410 core
out vec4 FragColor;


struct PointLight {
    vec3 position;
    
    float constant;
    float linear;
    float quadratic;
	vec3 lightColor;
    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
};



#define NR_POINT_LIGHTS 8 // maximun lights

in vec3 FragPos;
in vec3 Normal;
in vec2 TexCoord;

uniform int numLights; // total ligts from the pc program that will be rendered in gpu
uniform vec3 objectColor;
uniform vec3 viewPos;
uniform sampler2D texSampler;
uniform sampler2D texSampler2;
uniform PointLight pointLights[NR_POINT_LIGHTS];


// function prototypes

vec3 CalcPointLight(PointLight light, vec3 normal, vec3 fragPos, vec3 viewDir);


void main()
{    
    // properties
    vec3 norm = normalize(Normal);
    vec3 viewDir = normalize(viewPos - FragPos);
    
    // init value for result

    vec3 result = vec3(0.0,0.0,0.0);
    // sum of all ligts
    for(int i = 0; i < numLights; i++)
        result += CalcPointLight(pointLights[i], norm, FragPos, viewDir);    
    
    // textured if texture is not empty, else colored
    if (textureSize(texSampler, 0).x > 1){
       
        if (textureSize(texSampler2, 0).x > 1){
            FragColor = mix(texture(texSampler, TexCoord), texture(texSampler2, TexCoord) * vec4(result, 1.0f), 0.5);
        } else {
            FragColor = texture(texSampler, TexCoord) * vec4(result, 1.0);
        }
        
    }
    else {
        result = result * objectColor;
        FragColor = vec4(result, 1.0);
    }    
    
}


// calculates the color when using a point light.
vec3 CalcPointLight(PointLight light, vec3 normal, vec3 fragPos, vec3 viewDir)
{
    vec3 lightDir = normalize(light.position - fragPos);
    // diffuse shading
    float diff = max(dot(normal, lightDir), 0.0);
    // specular shading
    vec3 reflectDir = reflect(-lightDir, normal);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 32.0);
    // attenuation
    float pdistance = length(light.position - fragPos);
    float attenuation = 1.0 / (light.constant + light.linear * pdistance + light.quadratic * (pdistance * pdistance));    
    // combine results
    vec3 ambient = light.ambient;
    vec3 diffuse = light.diffuse * diff * light.lightColor;
    vec3 specular = light.specular * spec * light.lightColor;
    ambient *= attenuation;
    diffuse *= attenuation;
    specular *= attenuation;
    return (ambient + diffuse + specular);
}
