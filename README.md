# Go PDF Processor

Tool for **compressing PDF files** and **converting PDFs to Word (DOCX)** documents.

This project leverages an advanced pipeline using **Ghostscript**, **QPDF** and **Python** (PyMuPDF) to ensure maximum optimization and reliable text extraction. It features both a **Web Interface** and a **CLI**.

## 🚀 Features

### 📉 PDF Compression
The tool uses a multi-stage pipeline (Ghostscript → QPDF → Validation) to reduce file size without compromising readability.
- **Multiple Levels:**
  - `screen`: 72 dpi (Low quality, smallest size).
  - `ebook`: 150 dpi (Medium quality, balanced).
  - `printer`: 300 dpi (High quality).
  - `extreme`: Aggressive optimization (72 dpi, RGB conversion).

### 📝 PDF to Word Conversion
- **Linearized Output:** Converts complex layouts (like newspapers with columns) into a single column, top-to-bottom reading flow.
- **Text-Only Focus:** Automatically removes images and heavy graphics to prevent formatting errors and ensure the output is lightweight and easy to edit.
- **Robust:** Handles Cyrillic fonts and print-ready (CMYK) PDFs correctly.

---

### 📂 Project Structure
.
├── cmd/
│   ├── server/       # Entry point for the Web Server
│   └── cli/          # Entry point for the CLI tool
├── internal/
│   ├── api/          # HTTP Handlers and Router
│   ├── config/       # Env configuration loader
│   └── pdf/          # Core logic (Compressor, Converter, Scripts)
├── web/              # HTML Templates and Assets
├── test/             # Test files (e.g., newspaper.pdf)
└── Dockerfile        # Multi-stage Docker build

## 🐳 Getting Started (Docker Compose)

The easiest way to run the application is using Docker, as it automatically installs all external dependencies (Ghostscript, Python, QPDF).

### 1. Prerequisites
- Docker & Docker Compose installed.

### 2. Setup
Clone the repository and update the `.env` file (optional)

**⚙️ Configuration**
Variable	            Description	                                Default
PORT	                The HTTP port to bind to.	                8080
MAX_FILE_UPLOAD_SIZE	Max upload size in Megabytes (MB).	        50
CLEANUP_CRON_INTERVAL	How often (in minutes) to delete old files.	10

### 3. Start

Start the server: `docker compose up -d`

The Web UI will be accessible at: `http://localhost:8080`

To shut down the project use: `docker compose down`

### 3.1 💻 CLI Usage

You can use the tool via Command Line Interface (CLI) to process files in bulk.

**CLI Flags**
Flag	Description	                                    Default	    Values
- mode	Operation mode	                                `compress`	`compress`, `word`
- level	Compression level (only for compress mode)	    `ebook`	    `screen`, `ebook`, `printer`, `extreme`
- out	Output directory	                            uploads	    Any valid path
- sort  Enable smart sorting for columns (word only)    `true`      `true`, `false`

`docker compose run --rm app go run cmd/cli/main.go [flags] <files>`

**Examples**

1. Compress a file (Default - Ebook level):

`docker compose run --rm app go run cmd/cli/main.go -mode compress input.pdf`

2. Extreme Compression:

`docker compose run --rm app go run cmd/cli/main.go -mode compress -level extreme input.pdf`

3. Convert PDF to Word:

`docker compose run --rm app go run cmd/cli/main.go -mode word input.pdf`

`docker compose run --rm app go run cmd/cli/main.go -mode word input.pdf false`

### 4. 🧪 Running Tests

To run tests: `docker compose run --rm app go test ./... -v`
