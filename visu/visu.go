package visu

import (
	"audio-player/gtime"
	"bytes"
	"image"
	"image/png"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
)

// GetSize returns the size of the image returned by GenerateImage.
func GetSize() (int, int) {
	return 1920, 1080
}

// GenerateImage returns a waveform png image from the given audio file path.
func GenerateImage(path string) ([]byte, error) {
	gtime.Start("GenerateImage")

	h := getHash(path)

	bits, err := getImageByHash(h)
	if err == nil {
		return bits, nil
	}

	outPath := getImageOutPath(h)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	if _, exists := os.Stat(outPath); exists == nil {
		os.Remove(outPath)
	}

	width, height := GetSize()

	gtime.Start("GenerateImage.ffmpeg")
	cmd := exec.Command("ffmpeg", "-i", path, "-filter_complex", "showwavespic=s="+strconv.Itoa(width)+"x"+strconv.Itoa(height)+":colors=red:scale=sqrt:draw=full", "-frames:v", "1", outPath)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	gtime.End("GenerateImage.ffmpeg")

	file, err := os.Open(outPath)
	if err != nil {
		return nil, err
	}

	bits, err = io.ReadAll(file)
	if err != nil {
		file.Close()
		os.Remove(outPath)
		return nil, err
	}

	file.Close()

	gtime.Start("GenerateImage.Crop")
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

	//log.Println("transparency:", top, bottom)
	newImageHeight := bounds.Dy() - top - bottom

	newImage := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), newImageHeight))
	for y := 0; y < newImageHeight; y++ {
		for x := 0; x < bounds.Dx(); x++ {
			newImage.Set(x, y, img.At(x, y+top))
		}
	}

	gtime.End("GenerateImage.Crop")

	// convert newImage to png byte array
	var buf bytes.Buffer
	err = png.Encode(&buf, newImage)
	if err != nil {
		return nil, err
	}

	// finally, remove old image and save new one.
	if err := os.Remove(outPath); err != nil {
		log.Println("Failed to remove old image:", err)
	}

	if err := os.WriteFile(outPath, buf.Bytes(), 0600); err != nil {
		log.Println("Failed to write new image:", err)
	}

	gtime.End("GenerateImage")

	return buf.Bytes(), nil
}
