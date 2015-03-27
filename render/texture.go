package render

import (
	"fmt"
	"image"
	"image/png"
	"sync"

	"github.com/thinkofdeath/steven/render/atlas"
	"github.com/thinkofdeath/steven/resource"
)

var (
	textures    []*atlas.Type
	textureMap  = map[string]TextureInfo{}
	textureLock sync.RWMutex
)

// TextureInfo returns information about a texture in an atlas
type TextureInfo struct {
	Atlas int
	*atlas.Rect
}

// GetTexture returns the related TextureInfo for the requested texture.
// If the texture isn't found a placeholder is returned instead.
func GetTexture(name string) TextureInfo {
	textureLock.RLock()
	defer textureLock.RUnlock()
	t, ok := textureMap[name]
	if !ok {
		return textureMap["missing_texture"]
	}
	return t
}

// TODO(Think) better error handling (if possible to recover?)
// TODO(Think) Store textures
func loadTextures() {
	textureLock.Lock()
	defer textureLock.Unlock()

	// Clear existing
	textures = nil
	textureMap = map[string]TextureInfo{}

	for _, file := range resource.Search("minecraft", "textures/blocks/", ".png") {
		r, err := resource.Open("minecraft", file)
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(r)
		if err != nil {
			panic(err)
		}
		width, height := img.Bounds().Dx(), img.Bounds().Dy()
		if width != height {
			fmt.Printf("Skipping %s for now...\n", file)
			continue
		}
		var pix []byte
		switch img := img.(type) {
		case *image.NRGBA:
			pix = img.Pix
		case *image.RGBA:
			pix = img.Pix
		default:
			panic(fmt.Sprintf("unsupported image type %T", img))
		}
		name := file[len("textures/blocks/") : len(file)-4]
		at, rect := addTexture(pix, width, height)
		textureMap[name] = TextureInfo{
			Rect:  rect,
			Atlas: at,
		}
	}

	at, rect := addTexture([]byte{
		0, 0, 0, 255,
		255, 0, 255, 255,
		0, 0, 0, 255,
		255, 0, 255, 255,
	}, 2, 2)
	textureMap["missing_texture"] = TextureInfo{
		Rect:  rect,
		Atlas: at,
	}
}

func addTexture(pix []byte, width, height int) (int, *atlas.Rect) {
	for i, a := range textures {
		rect, err := a.Add(pix, width, height)
		if err == nil {
			return i, rect
		}
	}
	a := atlas.New(1024, 1024, 4)
	textures = append(textures, a)
	rect, err := a.Add(pix, width, height)
	if err != nil {
		panic(err)
	}
	return len(textures) - 1, rect
}
