# This Dockerfile is now clean because it will be built on a native amd64 server in Cloud Build.
FROM golang:1.22-bullseye AS builder

WORKDIR /app

# Install standard dependencies. No cross-compilers needed.
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    libleptonica-dev \
    libtesseract-dev \
    pkg-config \
    gcc \
    g++

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# A standard build command. No CGO flags, no GOARCH, no platform specifics.
RUN CGO_ENABLED=1 go build -o /server .

# --- Final Image ---
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-chi-sim \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /server /server

ENV PORT=8080
EXPOSE 8080

CMD ["/server"]