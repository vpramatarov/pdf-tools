package handlers

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func (h *Handler) StartCleanupCron() {
	checkInterval := time.Duration(h.Cfg.CleanupIntervalMinutes) * time.Minute
	//const checkInterval = 10 * time.Minute

	ticker := time.NewTicker(checkInterval)

	go func() {
		for range ticker.C {
			log.Println("🧹 Starting cleanup of old files...")
			h.deleteOldFiles()
		}
	}()
}

func (h *Handler) deleteOldFiles() {
	files, err := os.ReadDir(h.Cfg.UploadDir)
	if err != nil {
		log.Printf("⚠️ Error reading upload dir: %v", err)
		return
	}

	deletedCount := 0

	for _, file := range files {
		fullPath := filepath.Join(h.Cfg.UploadDir, file.Name())
		err := os.Remove(fullPath)
		if err != nil {
			log.Printf("❌ Failed to delete %s: %v", file.Name(), err)
		} else {
			deletedCount++
		}
	}

	if deletedCount > 0 {
		log.Printf("✅ Cleanup finished. Deleted %d old files.", deletedCount)
	}
}
