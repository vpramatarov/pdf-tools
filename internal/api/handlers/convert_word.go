package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/vpramatarov/pdf-tools/internal/pdf"
)

func (h *Handler) ConvertToWord(w http.ResponseWriter, r *http.Request) {
	// Calculate the limit in bytes: MB * 1024 * 1024
	maxBytes := h.Cfg.MaxUploadSizeMB << 20 // bytes shifting << 20
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Invalid file or 'pdf' field missing", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempInput := filepath.Join(h.Cfg.UploadDir, fmt.Sprintf("word_in_%d_%s", time.Now().Unix(), handler.Filename))
	f, err := os.Create(tempInput)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	io.Copy(f, file)
	f.Close()

	defer os.Remove(tempInput)

	sortParam := r.FormValue("sort")
	useSort := true
	if sortParam == "false" || sortParam == "0" {
		useSort = false
	}

	converter := pdf.NewConverter()
	generatedPath, err := converter.ToWord(tempInput, h.Cfg.UploadDir, useSort)
	if err != nil {
		http.Error(w, "Conversion failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	outputName := filepath.Base(generatedPath)

	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
		<div class="p-4 bg-blue-100 border border-blue-400 text-blue-700 rounded fade-in">
			<div class="flex items-center mb-2">
				<svg class="w-6 h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
				<span class="font-bold text-lg">Success!</span>
			</div>
			
			<p class="mb-4 text-sm">Your Word document is ready.</p>

			<a href="/download/%s" 
			   class="block w-full text-center text-white bg-blue-600 hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5">
			   ⬇️ Download .docx
			</a>
		</div>
	`, outputName)

	w.Write([]byte(html))
}
