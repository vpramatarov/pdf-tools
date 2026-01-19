package pdf

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

const testFileRelativePath = "../../test/newspaper.pdf"

func setupTestFile(t *testing.T) (string, string) {
	if _, err := os.Stat(testFileRelativePath); os.IsNotExist(err) {
		t.Fatalf("❌ Test file not found at: %s. \nMake sure you have 'test/newspaper.pdf' in the project root.", testFileRelativePath)
	}

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "newspaper.pdf")

	srcFile, err := os.Open(testFileRelativePath)
	if err != nil {
		t.Fatalf("Failed to open source file: %v", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	return tempDir, destPath
}
