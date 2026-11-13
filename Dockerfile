FROM golang:1.27.4-alpine3.24 AS builder

RUN apk add --no-cache curl

WORKDIR /app

COPY src src
COPY templates templates
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download
RUN go build -x -o media-roller ./src

# yt-dlp needs python
FROM python:3.15.0a8-alpine3.23

# This is where the downloaded files will be saved in the container.
ENV MR_DOWNLOAD_DIR="/download"

RUN apk add --update --no-cache \
    # https://github.com/yt-dlp/yt-dlp/issues/14404 \
    deno \
    curl \
    gcc \
    musl-dev \
    python3-dev \
    libffi-dev

# https://hub.docker.com/r/mwader/static-ffmpeg/tags
# https://github.com/wader/static-ffmpeg
COPY --from=mwader/static-ffmpeg:9.0.2-1 /ffmpeg  /usr/local/bin/
COPY --from=mwader/static-ffmpeg:9.0.2-1 /ffprobe /usr/local/bin/
COPY --from=builder /app/media-roller /app/media-roller
COPY templates /app/templates
COPY static /app/static

WORKDIR /app

RUN pip install curl_cffi --upgrade --pre
RUN pip install -U yt-dlp-ejs

# Get new releases here https://github.com/yt-dlp/yt-dlp/releases
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/download/2026.10.26/yt-dlp -o /usr/local/bin/yt-dlp && \
    echo "e51e69e89489f199530f13d32a4342da3ddb2cf34842f364603a066db05afbf3 /usr/local/bin/yt-dlp" | sha256sum -c - && \
    chmod a+rx /usr/local/bin/yt-dlp

RUN yt-dlp --update --update-to nightly

# Sanity check
RUN yt-dlp --version && \
    ffmpeg -version

ENTRYPOINT ["/app/media-roller"]
