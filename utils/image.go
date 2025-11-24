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
	mu          sync.RWMutex
	rawCache    map[string]image.Image
	scaledCache map[string]vaxis.Image
	loading     map[string]bool
	vx          *vaxis.Vaxis
}

var ImageCache *GlobalImageCache
var once sync.Once

func InitImageCache(vx *vaxis.Vaxis) {
	once.Do(func() {
		ImageCache = &GlobalImageCache{
			rawCache:    make(map[string]image.Image),
			scaledCache: make(map[string]vaxis.Image),
			loading:     make(map[string]bool),
			vx:          vx,
		}
	})
}

func (c *GlobalImageCache) Get(url string, width, height int) (vaxis.Image, bool) {
	c.mu.RLock()

	cacheKey := fmt.Sprintf("%s|%d|%d", url, width, height)
	if img, ok := c.scaledCache[cacheKey]; ok {
		c.mu.RUnlock()
		return img, true
	}

	rawImg, hasRaw := c.rawCache[url]
	c.mu.RUnlock()

	if !hasRaw {
		c.LoadAsync(url)
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if img, ok := c.scaledCache[cacheKey]; ok {
		return img, true
	}

	vxImage, err := c.vx.NewImage(rawImg)
	if err != nil {
		log.Printf("Error creating vaxis image from cached raw %s: %v", url, err)
		return nil, false
	}

	originalBounds := rawImg.Bounds()
	originalWidth, originalHeight := originalBounds.Dx(), originalBounds.Dy()

	scaleFactor := math.Max(float64(width)/float64(originalWidth), float64(height)/float64(originalHeight))
	scaledWidth := int(float64(originalWidth) * scaleFactor)
	scaledHeight := int(float64(originalHeight) * scaleFactor)

	vxImage.Resize(scaledWidth, scaledHeight)

	c.scaledCache[cacheKey] = vxImage

	return vxImage, true
}

func (c *GlobalImageCache) LoadAsync(url string) {
	c.mu.Lock()
	if c.loading[url] {
		c.mu.Unlock()
		return
	}
	if _, ok := c.rawCache[url]; ok {
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
			c.vx.PostEvent(vaxis.Redraw{})
		}()

		img, err := DownloadImage(url)
		if err != nil {
			log.Printf("Error downloading image %s: %v", url, err)
			return
		}

		c.mu.Lock()
		c.rawCache[url] = img
		c.mu.Unlock()
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
