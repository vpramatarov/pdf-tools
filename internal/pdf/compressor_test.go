package pdf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCompressor_Compress_Integration(t *testing.T) {
	if _, err := exec.LookPath("gs"); err != nil {
		t.Skip("Ghostscript (gs) not found, skipping compression test")
	}

	tempDir, inputPath := setupTestFile(t)
	outputPath := filepath.Join(tempDir, "output_compressed.pdf")

	comp := NewCompressor()

	t.Logf("🚀 Starting compression test on: %s", inputPath)
	err := comp.Compress(inputPath, outputPath, LevelScreen)
	if err != nil {
		t.Errorf("Compress function returned error: %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Fatal("❌ Output file was not created!")
	}

	if info.Size() == 0 {
		t.Error("❌ Output file is empty (0 bytes)")
	}

	t.Logf("✅ Compression successful. Output size: %d bytes", info.Size())
}
