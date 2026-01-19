package pdf

import (
	"log"
	"os"
)

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func logSize(label string, path string) {
	info, err := os.Stat(path)
	if err != nil {
		log.Printf("   %s: [file not found]", label)
		return
	}

	sizeMB := float64(info.Size()) / (1024 * 1024)
	log.Printf("   📊 %s: %.2f MB (%d bytes)", label, sizeMB, info.Size())
}
