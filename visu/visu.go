package visu

import (
	"bytes"
	"image"
	"image/png"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/exec"
)

func GenerateImage(path string) ([]byte, error) {
	// ffmpeg -i "file.wav" -filter_complex "showwavespic=s=1920x1080:colors=blue" -frames:v 1 out.png
	// TODO: not multi process safe
	outPath := "out.png"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	if _, exists := os.Stat(outPath); exists == nil {
		os.Remove(outPath)
	}

	cmd := exec.Command("ffmpeg", "-i", path, "-filter_complex", "showwavespic=s=1920x1080:colors=red:scale=sqrt:draw=full", "-frames:v", "1", outPath)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(outPath)
	if err != nil {
		return nil, err
	}

	bits, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		os.Remove(outPath)
		return nil, err
	}

	file.Close()

	// now, crop the image
	img, _, err := image.Decode(bytes.NewReader(bits))
	if err != nil {
		return nil, err
	}
	// now, count the number of top and bottom rows that are transparent, then crop the top and bottom.
	bounds := img.Bounds()

	var top, bottom int
	for y := 0; y < bounds.Dy(); y++ {
		transparent := true
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0 {
				transparent = false
				break
			}
		}
		if !transparent {
			break
		}
		top++
	}
	for y := bounds.Dy() - 1; y >= 0; y-- {
		transparent := true
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0 {
				transparent = false
				break
			}
		}
		if !transparent {
			break
		}
		bottom++
	}

	log.Println("transparency:", top, bottom)
	newImageHeight := bounds.Dy() - top - bottom

	newImage := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), newImageHeight))
	for y := 0; y < newImageHeight; y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newImage.Set(x, y, img.At(x, y+top))
		}
	}

	// convert newImage to png byte array
	var buf bytes.Buffer
	err = png.Encode(&buf, newImage)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
