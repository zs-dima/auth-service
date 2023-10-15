package image_tool

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"

	blurhash "github.com/bbrks/go-blurhash"

	"golang.org/x/image/draw"
)

func ToAvatar(srcPhoto []byte) ([]byte, string, error) {
	if len(srcPhoto) == 0 {
		return nil, "", nil
	}

	img, _, err := image.Decode(bytes.NewReader(srcPhoto))
	if err != nil {
		return nil, "", fmt.Errorf("unable to decode image: %w", err)
	}

	blurHash, err := blurhash.Encode(4, 4, img)
	if err != nil {
		return nil, "", fmt.Errorf("unable to encode blurhash: %w", err)
	}

	avatar := Scale(img, 100, 100, draw.ApproxBiLinear)

	// Convert image.Image to []byte
	var buf bytes.Buffer
	err = png.Encode(&buf, avatar)
	if err != nil {
		return nil, "", fmt.Errorf("unable to encode image: %w", err)
	}

	return buf.Bytes(), blurHash, nil
}

// Scale scales an image to fit within the specified rectangle maintaining aspect ratio.
func Scale(src image.Image, desiredWidth, desiredHeight int, scaler draw.Scaler) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	var newWidth, newHeight int

	if desiredWidth > 0 {
		// Calculate height based on aspect ratio
		newWidth = desiredWidth
		newHeight = srcHeight * newWidth / srcWidth
	} else if desiredHeight > 0 {
		// Calculate width based on aspect ratio
		newHeight = desiredHeight
		newWidth = srcWidth * newHeight / srcHeight
	} else {
		// No dimensions provided, return the original image
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	scaler.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// Scale scales an image to fit within the specified rectangle.
func ScaleFit(src image.Image, rect image.Rectangle, scale draw.Scaler) image.Image {
	dst := image.NewRGBA(rect)
	scale.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}
