package main

import (
	"fmt"
	"image"
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

func main() {
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 720, "dicom")
	rl.SetWindowState(rl.FlagVsyncHint)
	rl.SetTargetFPS(60)

	imgs := getImages()
	fmt.Printf("num images: %v\n", len(imgs))

	textures, unload := GetTextures(imgs)
	defer unload()

	testimg := rl.LoadTexture("texture.png")
	for i := range len(textures) {
		textures[i] = testimg
	}

	// setup
	camera := createCamera()
	s := state{
		model:    rl.LoadModel("rect.glb"),
		pos:      rl.NewVector3(0, 0, 0),
		textures: textures,
		index:    0,
		timer:    0,
	}

	lastTime := time.Now()
	for !rl.WindowShouldClose() {
		// update
		now := time.Now()
		deltaTime := now.Sub(lastTime).Milliseconds()
		lastTime = now
		s.timer += deltaTime
		if s.timer > 60 {
			s.timer -= 60
			s.index += 1
			if s.index == len(s.textures) {
				s.index = 0
			}
		}
		//draw
		draw(camera, s)
	}

	rl.CloseWindow()
}

func draw(camera rl.Camera3D, state state) {

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	rl.BeginMode3D(camera)

	pos := rl.NewVector3(0, 0, 0)
	for i := range state.index {
		texture := state.textures[i]
		DrawRectTexture(texture, pos, 1, 1, rl.White)
		pos.Y = pos.Y + 0.01
	}
	// just a single slice
	rl.DrawModelEx(state.model, state.pos, rl.NewVector3(0, 1, 0), 0, rl.NewVector3(1, 1, 1), rl.White)
	rl.DrawGrid(10, 1.0)

	rl.EndMode3D()
	rl.EndDrawing()
}

func createCamera() rl.Camera3D {
	camera := rl.Camera3D{}
	camera.Target = rl.NewVector3(0.0, 0.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective
	camera.Position = rl.NewVector3(2, 2, 2)
	rl.UpdateCamera(&camera, rl.CameraCustom)
	return camera
}

type state struct {
	model    rl.Model
	pos      rl.Vector3
	textures []rl.Texture2D
	index    int
	timer    int64
}

func getImages() []image.Image {
	imgs := []image.Image{}
	for i := range 64 {
		dataset, err := dicom.ParseFile(fmt.Sprintf("data/mr-brain-upenn-gbm-series/instance-00%02d.dcm", i+1), nil)
		if err != nil {
			log.Fatal(err)
		}
		pixelDataElement, err := dataset.FindElementByTag(tag.PixelData)
		if err != nil {
			log.Fatal(err)
		}
		pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)
		for _, fr := range pixelDataInfo.Frames {
			img, _ := fr.GetImage()
			imgs = append(imgs, img)
		}
	}

	return imgs
}

func DrawRectTexture(texture rl.Texture2D, position rl.Vector3, width, length float32, color rl.Color) {
	x := position.X
	y := position.Y
	z := position.Z

	// Set desired texture to be enabled while drawing following vertex data
	defaultID := rl.GetTextureIdDefault()
	rl.SetTexture(texture.ID)

	// Vertex data transformation can be defined with the commented lines,
	// but in this example we calculate the transformed vertex data directly when calling rlVertex3f()
	// rl.PushMatrix()
	// NOTE: Transformation is applied in inverse order (scale -> rotate -> translate)
	//rl.Translatef(2.0, 0.0, 0.0)
	//rl.Rotatef(45, 0, 1, )
	//rl.Scalef(2.0, 2.0, 2.0)

	rl.Begin(rl.Quads)
	rl.Color4ub(color.R, color.G, color.B, color.A)
	// Top Face
	rl.Normal3f(0.0, 1.0, 0.0) // Normal Pointing Up
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y, z-length/2) // Top Left Of The Texture and Quad.
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y, z+length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y, z-length/2) // Top Right Of The Texture and Quad Bottom Face
	rl.Normal3f(0.0, -1.0, 0.0)           // Normal Pointing Down
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y, z-length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y, z-length/2) // Top Left Of The Texture and Quad
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y, z+length/2) // Bottom Right Of The Texture and Quad

	rl.End()
	//rl.PopMatrix()

	rl.SetTexture(defaultID)
}

func GetTextures(images []image.Image) ([]rl.Texture2D, func()) {
	textures := []rl.Texture2D{}
	for _, k := range images {
		textures = append(textures, rl.LoadTextureFromImage(rl.NewImageFromImage(k)))
	}
	return textures, func() {
		for _, k := range textures {
			rl.UnloadTexture(k)
		}
	}
}
