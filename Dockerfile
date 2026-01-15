# Stage 1: Builder
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# Stage 2: Runtime (Production)
FROM debian:bookworm-slim AS runner

# Install system dependencies
# - ghostscript & qpdf for compression
# - python3 & pip for word conversion
# - --no-install-recommends saves space
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    ghostscript \
    qpdf \
    python3 \
    python3-pip \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

RUN pip3 install pymupdf python-docx --break-system-packages

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/web ./web

# Python scripts
COPY internal/pdf/scripts/ ./internal/pdf/scripts/

EXPOSE 8080
CMD ["./main"]

# Stage 3: Development (with Air)
FROM golang:1.25 AS dev

# Dev dependencies
RUN apt-get update && apt-get install -y \
    ghostscript \
    qpdf \
    python3 \
    python3-pip \
    git

RUN pip3 install pymupdf python-docx --break-system-packages

RUN go install github.com/air-verse/air@latest

WORKDIR /app
COPY go.mod go.sum ./
COPY test ./test
RUN go mod download
COPY . .

CMD ["air", "-c", ".air.toml"]