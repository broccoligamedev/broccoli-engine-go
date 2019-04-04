package main

import (
	"errors"
	"github.com/go-gl/gl/v3.3-core/gl"
	mathgl "github.com/go-gl/mathgl/mgl32"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var defaultShader *Shader

func InitGraphics() {
	var err error
	defaultShader, err = NewShader("shaders/default-vert.glsl", "shaders/default-frag.glsl")
	if err != nil {
		panic(err)
	}
	log.Println("OpenGL Version", GetOpenGLVersion())
	var data int32
	gl.GetIntegerv(gl.MAX_VERTEX_ATTRIBS, &data)
	log.Println("Max Vertex Attribs: ", data)
	gl.Enable(gl.POLYGON_SMOOTH)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.BLEND)
}

// ShaderID ...
type ShaderID uint32

// ProgramID ...
type ProgramID uint32

// BufferID ...
type BufferID uint32

// TextureID ...
type TextureID uint32

// Shader ...
type Shader struct {
	Id              ProgramID
	VertexPath      string
	FragmentPath    string
	vertexModTime   time.Time
	fragmentModTime time.Time
}

func GenTexture() TextureID {
	var tex uint32
	gl.GenTextures(1, &tex)
	return TextureID(tex)
}

func BindTexture(tex TextureID) {
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
}

func GenBindTexture() TextureID {
	tex := GenTexture()
	BindTexture(tex)
	return tex
}

func (shader *Shader) CheckForChanges() {
	vertexModTime, err := GetModifiedTime(shader.VertexPath)
	check(err)
	fragModTime, err := GetModifiedTime(shader.FragmentPath)
	check(err)
	if vertexModTime.After(shader.vertexModTime) || fragModTime.After(shader.fragmentModTime) {
		id, err := CreateProgram(shader.VertexPath, shader.FragmentPath)
		if err != nil {
			log.Println(err)
		} else {
			gl.DeleteProgram(uint32(shader.Id))
			shader.Id = id
		}
	}
}

func GetModifiedTime(filepath string) (time.Time, error) {
	file, err := os.Stat(filepath)
	if err != nil {
		return time.Time{}, err
	}
	return file.ModTime(), nil
}

func LoadShader(path string, shaderType uint32) (ShaderID, error) {
	shaderFile, err := ioutil.ReadFile(path)
	check(err)
	shaderFileStr := string(shaderFile)
	shaderID, err := CreateShader(shaderFileStr, shaderType)
	if err != nil {
		return 0, err
	}
	return shaderID, nil
}

func NewShader(vertexPath string, fragmentPath string) (*Shader, error) {
	id, err := CreateProgram(vertexPath, fragmentPath)
	if err != nil {
		return nil, err
	}
	vertexModifiedTime, err := GetModifiedTime(vertexPath)
	if err != nil {
		return nil, err
	}
	fragmentModifiedTime, err := GetModifiedTime(fragmentPath)
	if err != nil {
		return nil, err
	}
	newShader := &Shader{
		Id:              id,
		VertexPath:      vertexPath,
		FragmentPath:    fragmentPath,
		vertexModTime:   vertexModifiedTime,
		fragmentModTime: fragmentModifiedTime,
	}
	return newShader, nil
}

func (shader *Shader) Use() {
	UseProgram(shader.Id)
}

func GetOpenGLVersion() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}

func CreateShader(shaderSource string, shaderType uint32) (ShaderID, error) {
	shaderID := gl.CreateShader(shaderType)
	shaderSource += "\x00" // null terminator for shader
	csource, free := gl.Strs(shaderSource)
	gl.ShaderSource(shaderID, 1, csource, nil)
	free()
	gl.CompileShader(shaderID)
	var status int32
	gl.GetShaderiv(shaderID, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shaderID, gl.INFO_LOG_LENGTH, &logLength)
		infoLog := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shaderID, logLength, nil, gl.Str(infoLog))
		log.Println(shaderSource)
		log.Println(infoLog)
		return 0, errors.New("Failed to compile shader")
	}
	return ShaderID(shaderID), nil
}

func CreateProgram(vertPath string, fragPath string) (ProgramID, error) {
	vert, err := LoadShader(vertPath, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	frag, err := LoadShader(fragPath, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	programID := gl.CreateProgram()
	gl.AttachShader(programID, uint32(vert))
	gl.AttachShader(programID, uint32(frag))
	gl.LinkProgram(programID)
	var success int32
	gl.GetProgramiv(programID, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(programID, gl.INFO_LOG_LENGTH, &logLength)
		infoLog := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(programID, logLength, nil, gl.Str(infoLog))
		log.Println("Failed to link program: \n" + infoLog)
		return 0, errors.New("Failed to compile shader")
	}
	gl.DeleteShader(uint32(vert))
	gl.DeleteShader(uint32(frag))
	return ProgramID(programID), nil
}

func GenBuffer() BufferID {
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	return BufferID(buffer)
}

func BindBuffer(target uint32, buffer BufferID) {
	gl.BindBuffer(target, uint32(buffer))
}

func GenBindBuffer(target uint32) BufferID {
	buffer := GenBuffer()
	BindBuffer(target, buffer)
	return buffer
}

func BindVertexArray(buffer BufferID) {
	gl.BindVertexArray(uint32(buffer))
}

func GenVertexArray() BufferID {
	var buffer uint32
	gl.GenVertexArrays(1, &buffer)
	return BufferID(buffer)
}

func GenBindVertexArray() BufferID {
	buffer := GenVertexArray()
	BindVertexArray(buffer)
	return buffer
}

func GenEBO() BufferID {
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	return BufferID(buffer)
}

func BufferDataFloat32(target uint32, data []float32, usage uint32) {
	gl.BufferData(target, len(data)*4, gl.Ptr(data), usage)
}

func BufferDataUint32(target uint32, data []uint32, usage uint32) {
	gl.BufferData(target, len(data)*4, gl.Ptr(data), usage)
}

func UnbindVertexArray() {
	gl.BindVertexArray(0)
}

func UseProgram(programID ProgramID) {
	gl.UseProgram(uint32(programID))
}

func NullTerminatedString(s string) *uint8 {
	return gl.Str(s + "\x00")
}

func (shader *Shader) SetFloat32(name string, value float32) {
	location := gl.GetUniformLocation(uint32(shader.Id), NullTerminatedString(name))
	gl.Uniform1f(location, value)
}

func (shader *Shader) SetInt32(name string, value int32) {
	location := gl.GetUniformLocation(uint32(shader.Id), NullTerminatedString(name))
	gl.Uniform1i(location, value)
}

func (shader *Shader) SetFloat32Vec4(name string, f0, f1, f2, f3 float32) {
	location := gl.GetUniformLocation(uint32(shader.Id), NullTerminatedString(name))
	gl.Uniform4f(location, f0, f1, f2, f3)
}

func (shader *Shader) SetMat4(name string, mat4 mathgl.Mat4) {
	location := gl.GetUniformLocation(uint32(shader.Id), NullTerminatedString(name))
	data := [16]float32(mat4)
	gl.UniformMatrix4fv(location, 1, false, &data[0])
}

func LoadTextureAlpha(filename string) TextureID {
	file, err := os.Open(filename)
	check(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	check(err)
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}
	textureID := GenBindTexture()
	//gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.MIRRORED_REPEAT)
	//gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.MIRRORED_REPEAT)
	//gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	//gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return TextureID(textureID)
}

func LoadTexture(filename string) TextureID {
	file, err := os.Open(filename)
	check(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	check(err)
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	pixels := make([]byte, w*h*3)
	bIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
		}
	}
	textureID := GenBindTexture()
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(w), int32(h), 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return TextureID(textureID)
}

type sprite struct {
	vertices []float32
	width    float32
	height   float32
	x        float32
	y        float32
	r        float32
}

func newSprite(x, y, width, height, r float32) *sprite {
	vertices := []float32{
		0.0, 0.0,
		0.0, 1.0,
		1.0, 1.0,
		1.0, 0.0,
	}
	s := &sprite{
		vertices: vertices,
		x:        x,
		y:        y,
		r:        r,
		width:    width,
		height:   height,
	}
	return s
}

func (s *sprite) transform() mathgl.Mat4 {
	transform := mathgl.Ident4().Mul4(mathgl.Mat4([16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		s.x, s.y, 0, 1,
	}))
	transform = transform.Mul4(mathgl.Mat4([16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		s.width / 2, s.height / 2, 0, 1,
	}))
	transform = transform.Mul4(mathgl.HomogRotate3DZ(mathgl.DegToRad(s.r)))
	transform = transform.Mul4(mathgl.Mat4([16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		-s.width / 2, -s.height / 2, 0, 1,
	}))
	transform = transform.Mul4(mathgl.Mat4([16]float32{
		s.width, 0, 0, 0,
		0, s.height, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}))
	return transform
}

type spriteBatch struct {
	VBO     BufferID
	VAO     BufferID
	EBO     BufferID
	shader  *Shader
	sprites []*sprite
	indices []uint32
}

func (b *spriteBatch) init() {
	var err error
	b.shader, err = NewShader("shaders/triangle-vert.glsl", "shaders/triangle-frag.glsl")
	check(err)
	b.VAO = GenVertexArray()
	b.VBO = GenBuffer()
	b.EBO = GenBuffer()
	b.indices = []uint32{}
}

func (b *spriteBatch) addSprite(s *sprite) {
	baseIndices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}
	for _, i := range baseIndices {
		b.indices = append(b.indices, i+uint32(4*len(b.sprites)))
	}
	b.sprites = append(b.sprites, s)
}

func (b *spriteBatch) draw() {
	projection := mathgl.Ortho(0, 600, 600, 0, -1, 1)
	finalVertices := []float32{}
	if len(b.sprites) == 0 {
		return
	}
	//log.Println("#sprites", len(b.sprites))
	for _, currentSprite := range b.sprites {
		currentTransform := currentSprite.transform()
		vertexBuffer := []float32{}
		for v := 0; v < len(currentSprite.vertices); v += 2 {
			newVertex := mathgl.Vec4([4]float32{
				currentSprite.vertices[v],
				currentSprite.vertices[v+1],
				0.0,
				1.0})
			newVertex = currentTransform.Mul4x1(newVertex)
			for i := 0; i < 2; i++ {
				vertexBuffer = append(vertexBuffer, newVertex[i])
			}
		}
		for _, v := range vertexBuffer {
			finalVertices = append(finalVertices, v)
		}
	}
	//log.Println("#vertices: ", len(finalVertices)/2)
	BindVertexArray(b.VAO)
	BindBuffer(gl.ARRAY_BUFFER, b.VBO)
	BufferDataFloat32(gl.ARRAY_BUFFER, finalVertices, gl.STATIC_DRAW)
	BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.EBO)
	BufferDataUint32(gl.ELEMENT_ARRAY_BUFFER, b.indices, gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	b.shader.Use()
	b.shader.SetMat4("projection", projection)
	//gl.DrawArrays(gl.TRIANGLES, 0, int32(len(finalVertices)))
	gl.DrawElements(gl.TRIANGLES, int32(len(finalVertices)), gl.UNSIGNED_INT, gl.PtrOffset(0))
}

func DrawTriangle(x1, y1, x2, y2, x3, y3 float32) {
	vertices := []float32{
		x1, y1,
		x2, y2,
		x3, y3,
	}
	projection := mathgl.Ortho(0, 600, 600, 0, -1, 1)
	//VAO := GenBindVertexArray()
	VBO := GenBindBuffer(gl.ARRAY_BUFFER)
	BindBuffer(gl.ARRAY_BUFFER, VBO)
	BufferDataFloat32(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	defaultShader.Use()
	defaultShader.SetMat4("projection", projection)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

// SetDrawColor ...
func SetDrawColor(r, g, b, a float32) {
	defaultShader.SetFloat32Vec4("inColor", r, g, b, a)
}

// DrawRect ...
func DrawRectangle(x, y, width, height float32) {
	// TODO(ryan): redundant usage of projection matrix since it never changes for our purpose (except
	// during screen resizing).
	// TODO(ryan): we should be batching primitives too
	vertices := []float32{
		x, y,
		x + width, y,
		x + width, y + height,
		x, y + height,
	}
	indices := []uint32{
		0, 1, 3,
		1, 2, 3,
	}
	GenBindBuffer(gl.ELEMENT_ARRAY_BUFFER)
	BufferDataUint32(gl.ELEMENT_ARRAY_BUFFER, indices, gl.STATIC_DRAW)
	projection := mathgl.Ortho(0, 600, 600, 0, -1, 1)
	VBO := GenBindBuffer(gl.ARRAY_BUFFER)
	BindBuffer(gl.ARRAY_BUFFER, VBO)
	BufferDataFloat32(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	defaultShader.Use()
	defaultShader.SetMat4("projection", projection)
	gl.DrawElements(gl.TRIANGLES, int32(len(vertices)), gl.UNSIGNED_INT, gl.PtrOffset(0))
}

// ClearScreen ...
func ClearScreen(r, g, b float32) {
	gl.ClearColor(r, g, b, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}
