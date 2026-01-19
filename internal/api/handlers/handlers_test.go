package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vpramatarov/pdf-tools/internal/config"
)

// Helper for creating multipart request
func createMultipartRequest(t *testing.T, uri, paramName, filePath string) (*http.Request, string) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	_ = writer.WriteField("level", "ebook")

	writer.Close()

	req := httptest.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, filepath.Base(filePath)
}

func TestHandler_Compress_Integration(t *testing.T) {
	if _, err := exec.LookPath("gs"); err != nil {
		t.Skip("Ghostscript (gs) not found, skipping handler integration test")
	}

	testFilePath := filepath.Join("..", "..", "..", "test", "newspaper.pdf")

	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		testFilePath = filepath.Join("..", "..", "test", "newspaper.pdf")
		if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
			t.Fatalf("Test file not found at %s. Make sure 'newspaper.pdf' is in the 'test' folder in project root.", testFilePath)
		}
	}

	origInfo, err := os.Stat(testFilePath)
	if err != nil {
		t.Fatalf("Failed to stat original file: %v", err)
	}

	originalSize := origInfo.Size()
	tempDir := t.TempDir()

	testCfg := &config.Config{
		UploadDir:              tempDir,
		MaxUploadSizeMB:        50,
		CleanupIntervalMinutes: 10,
	}
	h := New(testCfg)

	req, filename := createMultipartRequest(t, "/compress", "pdf", testFilePath)
	rr := httptest.NewRecorder()

	t.Logf("🚀 Testing compression on file: %s (Size: %d bytes)", filename, originalSize)
	h.Compress(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		t.Logf("Response body: %s", rr.Body.String())
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Handler returned wrong content type: got %v want text/html", contentType)
	}

	if !strings.Contains(rr.Body.String(), "/download/") {
		t.Errorf("Handler response does not contain download link. Body snippet: %s", rr.Body.String()[:200])
	}

	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	var compressedPath string
	var compressedSize int64

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".pdf") {
			info, err := file.Info()
			if err == nil && info.Size() > 0 {
				compressedPath = filepath.Join(tempDir, file.Name())
				compressedSize = info.Size()
				break
			}
		}
	}

	if compressedPath == "" {
		t.Fatalf("❌ No generated PDF file found in output directory: %s", tempDir)
	}

	t.Logf("📂 Found output file: %s", filepath.Base(compressedPath))
	t.Logf("📊 Original Size: %d bytes | Compressed Size: %d bytes", originalSize, compressedSize)

	// Проверка за ефективност
	if compressedSize >= originalSize {
		t.Errorf("❌ Compression failed to reduce file size. Original: %d, Compressed: %d", originalSize, compressedSize)
	} else {
		reduction := 100.0 - (float64(compressedSize)/float64(originalSize))*100.0
		t.Logf("✅ Success! File reduced by %.2f%%", reduction)
	}
}

func TestHandler_Compress_NoFile(t *testing.T) {
	testCfg := &config.Config{
		UploadDir:              t.TempDir(),
		MaxUploadSizeMB:        50,
		CleanupIntervalMinutes: 10,
	}

	h := New(testCfg)

	req := httptest.NewRequest("POST", "/compress", nil)
	rr := httptest.NewRecorder()

	h.Compress(rr, req)

	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected error code, got %v", rr.Code)
	}
}

func TestHandler_Home(t *testing.T) {
	testCfg := &config.Config{
		UploadDir:              t.TempDir(),
		MaxUploadSizeMB:        50,
		CleanupIntervalMinutes: 10,
	}
	// dummy template
	os.MkdirAll("web/templates", 0755)
	os.WriteFile("web/templates/index.html", []byte("<html></html>"), 0644)
	defer os.RemoveAll("web") // cleanup

	h := New(testCfg)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	h.Home(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Home handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
}
