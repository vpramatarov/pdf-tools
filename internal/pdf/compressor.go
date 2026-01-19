package pdf

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type CompressionLevel string

const (
	LevelScreen  CompressionLevel = "/screen"  // 72 dpi
	LevelEbook   CompressionLevel = "/ebook"   // 150 dpi
	LevelPrinter CompressionLevel = "/printer" // 300 dpi
	LevelExtreme CompressionLevel = "extreme"
)

type Compressor struct{}

func NewCompressor() *Compressor {
	return &Compressor{}
}

func (c *Compressor) Compress(inputPath string, outputPath string, level CompressionLevel) error {
	gsOut := outputPath + ".gs.pdf"
	qpdfOut := outputPath + ".qpdf.pdf"
	defer os.Remove(gsOut)
	defer os.Remove(qpdfOut)

	logSize("Original", inputPath)

	// --- Ghostscript (Images + Rendering) ---
	log.Println("🔹 Step 1: Ghostscript (Image processing)...")
	if err := c.runGhostscript(inputPath, gsOut, level); err != nil {
		return fmt.Errorf("step 1 failed: %w", err)
	}

	logSize("After Ghostscript", gsOut)

	// --- QPDF (Structure, Objects and Metadata) ---
	// QPDF is best at "Object Stream" compression
	log.Println("🔹 Step 2: QPDF (Structural cleanup & Metadata removal)...")
	if err := c.runQpdf(gsOut, outputPath); err != nil {
		log.Printf("⚠️ QPDF failed: %v. Proceeding with GS output.", err)
		// Fallback: copy the result from GS to the QPDF variable
		copyFile(gsOut, outputPath)
	}

	logSize("After QPDF (Final)", outputPath)

	log.Println("✅ Compression pipeline finished.")
	return nil
}

func (c *Compressor) runGhostscript(input string, output string, level CompressionLevel) error {
	_, err := exec.LookPath("gs")
	if err != nil {
		return fmt.Errorf("ghostscript (gs) not found")
	}

	args := []string{
		"gs",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-dDetectDuplicateImages=true",
		"-dCompressFonts=true",
		"-dSubsetFonts=true",
		"-dColorImageResolution=150",
		"-dGrayImageResolution=150",
		"-dMonoImageResolution=150",
		"-dRemoveUnusedResources=true",
		"-dNumRenderingThreads=4",
		// We use Bicubic (better quality) or Subsample (worse quality but faster)
		"-dColorImageDownsampleType=/Bicubic",
		"-dDiscardPageThumbnails=true",  // Remove hidden images for viewing
		"-dDiscardPageAnnotations=true", // Remove notes and forms
		"-dPreserveAnnots=false",        // Forces annotation removal
		"-r150",                         // Forces lower resolution rendering for vectors
	}

	switch level {
	case LevelExtreme:
		args = append(args,
			"-dPDFSETTINGS=/screen",
			// We prohibit GS from passing already compressed images.
			// This causes GS to decode and re-encode them with our settings.
			"-dPassThroughJPEGImages=false",
			// Force Downsampling (resolution reduction)
			"-dDownsampleColorImages=true",
			"-dDownsampleGrayImages=true",
			"-dDownsampleMonoImages=true",
			// set 72 DPI
			"-dColorImageResolution=72",
			"-dGrayImageResolution=72",
			"-dMonoImageResolution=72",
			// Conversion to RGB (saves CMYK ink channels)
			"-sColorConversionStrategy=RGB",
			"-sProcessColorModel=DeviceRGB",
			// Force JPEG compression (DCTEncode)
			"-dAutoFilterColorImages=false",
			"-dAutoFilterGrayImages=false",
			"-dEncodeColorImages=true",
			"-dColorImageFilter=/DCTEncode",
			"-dGrayImageFilter=/DCTEncode",
			"-dDiscardBookmarks=true",
		)
	case LevelScreen:
		args = append(args,
			"-dPDFSETTINGS=/screen",
			"-dColorImageResolution=72",
			"-dGrayImageResolution=72",
			"-dMonoImageResolution=72",
			"-dDiscardBookmarks=true",
		)
	default:
		// allow PassThrough for better quality
		args = append(args, fmt.Sprintf("-dPDFSETTINGS=%s", level))
	}

	args = append(args, fmt.Sprintf("-sOutputFile=%s", output), input)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (c *Compressor) runQpdf(input string, output string) error {
	_, err := exec.LookPath("qpdf")
	if err != nil {
		return fmt.Errorf("qpdf not found")
	}

	args := []string{
		"qpdf",
		"--recompress-flate",        // recompresses all text streams
		"--object-streams=generate", // object grouping
		"--stream-data=compress",    // guarantees data compression
		"--compression-level=9",     // maximum compression
		input,
		output,
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
