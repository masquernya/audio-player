package visu

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"time"
)

var cachePath = path.Join(path.Dir(os.Args[0]), "cache")

func ClearCache() {
	files, err := os.ReadDir(cachePath)
	if err != nil {
		log.Println("Failed to read cache directory:", err)
		return
	}

	var size int64
	const maxCacheSize int64 = 1024 * 1024 * 1024 // 1GB
	// delete least recently used/created (I.e. sort by most recent)
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := files[i].Info()
		infoJ, _ := files[j].Info()
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			log.Println("Failed to get file info:", err)
			continue
		}
		size += info.Size()

		if size > maxCacheSize {
			err := os.Remove(path.Join(cachePath, file.Name()))
			if err != nil {
				log.Println("Failed to remove cache file:", err)
			}
		}
	}

}

func getHash(filePath string) string {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return ""
	}
	bits, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(bits)
	return hex.EncodeToString(hash[:])
}

func getImageByHash(h string) ([]byte, error) {
	if h == "" {
		return nil, os.ErrNotExist
	}

	fullPath := path.Join(cachePath, h+".png")
	bits, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	// update mod time
	if err := os.Chtimes(fullPath, time.Now(), time.Now()); err != nil {
		log.Println("Failed to update file mod time:", err)
	}
	return bits, nil
}

func getImageOutPath(h string) string {
	if h == "" {
		return ""
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		if err = os.Mkdir(cachePath, 0700); err != nil {
			log.Fatal("Failed to create cache directory:", err)
		}
	}

	return path.Join(cachePath, h+".png")
}
