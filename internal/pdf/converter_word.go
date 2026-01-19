package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) ToWord(inputPath string, outputDir string, sort bool) (string, error) {
	scriptPath, err := c.findScriptPath()
	if err != nil {
		return "", err
	}

	textOnlyPath := filepath.Join(filepath.Dir(inputPath), "clean_"+filepath.Base(inputPath))
	defer os.Remove(textOnlyPath)

	if err := c.removeImages(inputPath, textOnlyPath); err != nil {
		fmt.Printf("⚠️ GS cleanup failed: %v. Using original.\n", err)
		copyFile(inputPath, textOnlyPath)
	}

	// Python conversion
	fileName := filepath.Base(inputPath)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	docxPath := filepath.Join(outputDir, baseName+".docx")

	sortArg := "true"
	if !sort {
		sortArg = "false"
	}

	cmd := exec.Command("python3", scriptPath, textOnlyPath, docxPath, sortArg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("linearization error: %w", err)
	}

	return docxPath, nil
}

func (c *Converter) removeImages(input string, output string) error {
	cmd := exec.Command("gs",
		"-o", output,
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dFILTERIMAGE",  // Removes images
		"-dFILTERVECTOR", // Removes vectors
		input,
	)
	return cmd.Run()
}

func (c *Converter) findScriptPath() (string, error) {
	const scriptName = "convert_word.py"

	possiblePaths := []string{
		// Production: when run from root (./main)
		filepath.Join("internal", "pdf", "scripts", scriptName),

		// Testing: when run from internal/pdf (go test)
		filepath.Join("scripts", scriptName),

		// Absolute path (Docker Safe)
		"/app/internal/pdf/scripts/" + scriptName,
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("python script '%s' not found in any expected location. Checked: %v", scriptName, possiblePaths)
}
