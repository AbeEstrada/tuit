package utils

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"net/http"
	"sync"

	_ "golang.org/x/image/webp"

	"git.sr.ht/~rockorager/vaxis"
)

type GlobalImageCache struct {
	mu      sync.RWMutex
	cache   map[string]vaxis.Image
	loading map[string]bool
	vx      *vaxis.Vaxis
}

var ImageCache *GlobalImageCache
var once sync.Once

func InitImageCache(vx *vaxis.Vaxis) {
	once.Do(func() {
		ImageCache = &GlobalImageCache{
			cache:   make(map[string]vaxis.Image),
			loading: make(map[string]bool),
			vx:      vx,
		}
	})
}

func (c *GlobalImageCache) Get(url string) (vaxis.Image, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	img, found := c.cache[url]
	return img, found
}

func (c *GlobalImageCache) LoadAsync(url string, width, height int) {
	c.mu.Lock()
	if c.loading[url] {
		c.mu.Unlock()
		return
	}
	c.loading[url] = true
	c.mu.Unlock()

	go func() {
		defer func() {
			c.mu.Lock()
			delete(c.loading, url)
			c.mu.Unlock()
		}()

		img, err := DownloadImage(url)
		if err != nil {
			log.Printf("Error downloading image %s: %v", url, err)
			return
		}

		vxImage, err := c.vx.NewImage(img)
		if err != nil {
			log.Printf("Error creating vaxis image from %s: %v", url, err)
			return
		}

		originalBounds := img.Bounds()
		originalWidth, originalHeight := originalBounds.Dx(), originalBounds.Dy()

		scaleFactor := math.Max(float64(width)/float64(originalWidth), float64(height)/float64(originalHeight))
		scaledWidth := int(float64(originalWidth) * scaleFactor)
		scaledHeight := int(float64(originalHeight) * scaleFactor)

		vxImage.Resize(scaledWidth, scaledHeight)

		c.mu.Lock()
		c.cache[url] = vxImage
		c.mu.Unlock()

		c.vx.PostEvent(vaxis.Redraw{})
	}()
}

func DownloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get image: bad status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return img, nil
}
