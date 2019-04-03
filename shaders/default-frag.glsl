#version 330 core

out vec4 fragColor;
uniform vec4 inColor;

void main() {
    fragColor = vec4(inColor.r, inColor.g, inColor.b, 0);
}