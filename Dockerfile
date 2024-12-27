FROM golang:1.23.4-alpine3.21 AS builder

RUN apk add --no-cache curl

WORKDIR /app

COPY src src
COPY templates templates
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download
RUN go build -x -o media-roller ./src

# yt-dlp needs python
FROM python:3.13.1-alpine3.21

# This is where the downloaded files will be saved in the container.
ENV MR_DOWNLOAD_DIR="/download"

RUN apk add --update --no-cache \
    curl

# https://hub.docker.com/r/mwader/static-ffmpeg/tags
# https://github.com/wader/static-ffmpeg
COPY --from=mwader/static-ffmpeg:7.1 /ffmpeg  /usr/local/bin/
COPY --from=mwader/static-ffmpeg:7.1 /ffprobe /usr/local/bin/
COPY --from=builder /app/media-roller /app/media-roller
COPY templates /app/templates
COPY static /app/static

WORKDIR /app

# Get new releases here https://github.com/yt-dlp/yt-dlp/releases
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/download/2024.12.13/yt-dlp -o /usr/local/bin/yt-dlp && \
    echo "ec5f59f8b8908d93b2bdf6663c3ecba70781f39de21255d183220f250ebccc94 /usr/local/bin/yt-dlp" | sha256sum -c - && \
    chmod a+rx /usr/local/bin/yt-dlp

RUN yt-dlp --update --update-to nightly

# Sanity check
RUN yt-dlp --version && \
    ffmpeg -version

ENTRYPOINT ["/app/media-roller"]