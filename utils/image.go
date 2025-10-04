package utils

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
)

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
