package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/vpramatarov/pdf-tools/internal/pdf"
)

func (h *Handler) Compress(w http.ResponseWriter, r *http.Request) {
	// Calculate the limit in bytes: MB * 1024 * 1024
	maxBytes := h.Cfg.MaxUploadSizeMB << 20 // bytes shifting << 20
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "File too large or invalid form", http.StatusBadRequest)
		return
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["pdf"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	results := make([]processingResult, len(files))
	var wg sync.WaitGroup
	var mu sync.Mutex

	levelStr := r.FormValue("level")
	level := pdf.LevelEbook
	switch levelStr {
	case "extreme":
		level = pdf.LevelExtreme
	case "screen":
		level = pdf.LevelScreen
	case "printer":
		level = pdf.LevelPrinter
	}

	compressor := pdf.NewCompressor()

	for i, fileHeader := range files {
		wg.Add(1)
		go func(idx int, fh *multipart.FileHeader) {
			defer wg.Done()

			srcFile, err := fh.Open()
			if err != nil {
				return
			}
			defer srcFile.Close()

			tempInput := filepath.Join(h.Cfg.UploadDir, fmt.Sprintf("in_%d_%d_%s", time.Now().Unix(), idx, fh.Filename))
			dstFile, err := os.Create(tempInput)
			if err != nil {
				return
			}
			io.Copy(dstFile, srcFile)
			dstFile.Close()

			info, _ := os.Stat(tempInput)
			origSize := info.Size()

			tempOutput := filepath.Join(h.Cfg.UploadDir, fmt.Sprintf("compressed_%d_%d_%s", time.Now().Unix(), idx, fh.Filename))
			err = compressor.Compress(tempInput, tempOutput, level)

			finalSize := int64(0)
			finalPath := tempOutput

			if err == nil {
				outInfo, _ := os.Stat(tempOutput)
				finalSize = outInfo.Size()

				// Revert if larger
				if finalSize >= origSize {
					finalSize = origSize
					inputContent, _ := os.ReadFile(tempInput)
					os.WriteFile(tempOutput, inputContent, 0644)
				}
			}

			os.Remove(tempInput)

			mu.Lock()
			results[idx] = processingResult{
				originalPath:   tempInput,
				compressedPath: finalPath,
				originalSize:   origSize,
				finalSize:      finalSize,
				filename:       fh.Filename,
				err:            err,
			}
			mu.Unlock()
		}(i, fileHeader)
	}

	wg.Wait()

	var totalOrig, totalFinal int64
	var finalDownloadName, displayTitle string

	validResults := []processingResult{}
	for _, res := range results {
		if res.err == nil && res.finalSize > 0 {
			totalOrig += res.originalSize
			totalFinal += res.finalSize
			validResults = append(validResults, res)
		}
	}

	if len(validResults) == 0 {
		http.Error(w, "Failed to compress files", http.StatusInternalServerError)
		return
	}

	if len(validResults) == 1 {
		res := validResults[0]
		finalDownloadName = filepath.Base(res.compressedPath)
		displayTitle = res.filename
	} else {
		zipName := fmt.Sprintf("compressed_batch_%d.zip", time.Now().Unix())
		zipPath := filepath.Join(h.Cfg.UploadDir, zipName)

		if err := createZip(zipPath, validResults); err != nil {
			http.Error(w, "Failed to create zip", http.StatusInternalServerError)
			return
		}
		finalDownloadName = zipName
		displayTitle = fmt.Sprintf("Archive created from %d files", len(validResults))
	}

	savedBytes := totalOrig - totalFinal
	savedPercent := 0.0
	if totalOrig > 0 {
		savedPercent = (float64(savedBytes) / float64(totalOrig)) * 100
	}

	statusColor := "green"
	if savedBytes <= 0 {
		statusColor = "yellow"
	}

	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
		<div class="p-4 bg-%s-100 border border-%s-400 text-%s-700 rounded fade-in">
			<div class="flex items-center mb-2">
				<svg class="w-6 h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
				<span class="font-bold text-lg">%s</span>
			</div>
			
			<div class="grid grid-cols-2 gap-4 text-sm mb-4">
				<div>
					<p class="text-gray-600">Original total size:</p>
					<p class="font-semibold">%s</p>
				</div>
				<div>
					<p class="text-gray-600">New total size:</p>
					<p class="font-semibold">%s</p>
				</div>
			</div>

			<div class="mb-4">
				<div class="w-full bg-gray-200 rounded-full h-2.5">
					<div class="bg-%s-600 h-2.5 rounded-full" style="width: %.0f%%"></div>
				</div>
				<p class="text-xs text-right mt-1 font-bold">Saved total: %s (%.1f%%)</p>
			</div>

			<a href="/download/%s" 
			   class="block w-full text-center text-white bg-%s-600 hover:bg-%s-700 focus:ring-4 focus:ring-%s-300 font-medium rounded-lg text-sm px-5 py-2.5">
			   ⬇️ Download (%s)
			</a>
		</div>
	`,
		statusColor, statusColor, statusColor,
		displayTitle,
		formatSize(totalOrig),
		formatSize(totalFinal),
		statusColor, savedPercent,
		formatSize(savedBytes), savedPercent,
		finalDownloadName,
		statusColor, statusColor, statusColor,
		filepath.Ext(finalDownloadName))

	w.Write([]byte(html))
}
