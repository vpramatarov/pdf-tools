package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vpramatarov/pdf-tools/internal/pdf"
)

func main() {
	levelFlag := flag.String("level", "ebook", "Compression level: extreme, screen, ebook, printer")
	outDirFlag := flag.String("out", "uploads", "Output directory for compressed files")
	modeFlag := flag.String("mode", "compress", "Mode: 'compress' or 'word'")
	sortMode := flag.Bool("sort", true, "Enable smart sorting for columns (default true)")
	flag.Parse()
	files := flag.Args()

	if len(files) == 0 {
		fmt.Println("Usage: go run cmd/cli/main.go [options] <files>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDirFlag, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	compressor := pdf.NewCompressor()
	converter := pdf.NewConverter()

	absOutDir, _ := filepath.Abs(*outDirFlag)
	fmt.Printf("📂 Saving files to: %s\n", absOutDir)

	var wg sync.WaitGroup
	startTime := time.Now()

	for _, inputFile := range files {
		wg.Add(1)
		go func(input string) {
			defer wg.Done()

			// --- Convert to WORD ---
			if *modeFlag == "word" {
				fmt.Printf("📝 Converting to Word: %s ...\n", filepath.Base(input))

				resPath, err := converter.ToWord(input, *outDirFlag, *sortMode)
				if err != nil {
					log.Printf("❌ Conversion failed for %s: %v", input, err)
					return
				}
				fmt.Printf("✅ Converted: %s\n", filepath.Base(resPath))
				return
			}

			// --- Compression (DEFAULT) ---
			baseName := filepath.Base(input)
			ext := filepath.Ext(input)
			newName := strings.TrimSuffix(baseName, ext) + "_compressed" + ext
			outputFile := filepath.Join(*outDirFlag, newName)

			var level pdf.CompressionLevel
			switch *levelFlag {
			case "extreme":
				level = pdf.LevelScreen
			case "screen":
				level = pdf.LevelScreen
			case "printer":
				level = pdf.LevelPrinter
			default:
				level = pdf.LevelEbook
			}

			fmt.Printf("⏳ Compressing %s ...\n", baseName)
			err := compressor.Compress(inputFile, outputFile, level)
			if err != nil {
				log.Printf("❌ Error compressing %s: %v", inputFile, err)
				return
			}
			checkSizeAndReport(inputFile, outputFile)
			fmt.Printf("Done: %s\n", outputFile)

		}(inputFile)
	}

	wg.Wait()
	fmt.Printf("\n✨ All done in %v\n", time.Since(startTime))
}

func checkSizeAndReport(input, output string) {
	inInfo, err1 := os.Stat(input)
	outInfo, err2 := os.Stat(output)

	if err1 != nil || err2 != nil {
		return
	}

	oldSize := inInfo.Size()
	newSize := outInfo.Size()

	if newSize >= oldSize {
		fmt.Printf("⚠️  %s: Result was larger/same, reverting to original.\n", filepath.Base(input))
		inputContent, _ := os.ReadFile(input)
		os.WriteFile(output, inputContent, 0644)
		newSize = oldSize
	}

	savedBytes := oldSize - newSize
	percent := 0.0
	if oldSize > 0 {
		percent = (float64(savedBytes) / float64(oldSize)) * 100
	}

	fmt.Printf("✅ %s: Saved %.1f%% (%s -> %s)\n",
		filepath.Base(input),
		percent,
		formatSize(oldSize),
		formatSize(newSize))
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
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
