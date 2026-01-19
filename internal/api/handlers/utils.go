package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

type processingResult struct {
	originalPath   string
	compressedPath string
	originalSize   int64
	finalSize      int64
	filename       string
	err            error
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func createZip(zipFilename string, results []processingResult) error {
	newZipFile, err := os.Create(zipFilename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, res := range results {
		f, err := os.Open(res.compressedPath)
		if err != nil {
			continue
		}

		w, err := zipWriter.Create(res.filename)
		if err != nil {
			f.Close()
			continue
		}

		if _, err := io.Copy(w, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}
