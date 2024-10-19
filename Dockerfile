FROM golang:1.23.2-alpine3.20 AS builder

RUN apk add --no-cache curl

WORKDIR /app

COPY src src
COPY templates templates
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download
RUN go build -x -o media-roller ./src

# yt-dlp needs python
FROM python:3.13.0-alpine3.20

# This is where the downloaded files will be saved in the container.
ENV MR_DOWNLOAD_DIR="/download"

RUN apk add --update --no-cache \
  curl

COPY --from=builder /app/media-roller /app/media-roller
COPY --from=mwader/static-ffmpeg:7.1 /ffmpeg /usr/local/bin/
COPY templates /app/templates
COPY static /app/static

WORKDIR /app

RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
   chmod a+rx /usr/local/bin/yt-dlp

# Sanity check
RUN yt-dlp --version && \
    ffmpeg -version

CMD /app/media-roller
