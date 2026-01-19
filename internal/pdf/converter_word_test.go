package pdf

import (
	"os"
	"os/exec"
	"testing"
)

func TestConverter_ToWord_Integration(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not found, skipping conversion test")
	}

	scriptPath := "./scripts/convert_word.py"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skipf("Script not found at %s, skipping test", scriptPath)
	}

	tempDir, inputPath := setupTestFile(t)

	conv := NewConverter()

	t.Logf("🚀 Starting PDF to Word conversion test on: %s", inputPath)

	docxPath, err := conv.ToWord(inputPath, tempDir, false)
	if err != nil {
		t.Errorf("ToWord function returned error: %v", err)
	}

	info, err := os.Stat(docxPath)
	if os.IsNotExist(err) {
		t.Fatal("❌ DOCX file was not created!")
	}

	if info.Size() < 500 {
		t.Errorf("❌ DOCX file seems too small (%d bytes)", info.Size())
	}

	t.Logf("✅ Conversion successful. Created: %s (%d bytes)", docxPath, info.Size())
}
