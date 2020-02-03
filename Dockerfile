FROM golang:1.13.6-alpine3.11 as builder

RUN apk add --no-cache curl

WORKDIR /app

COPY src src
COPY templates templates
COPY go.mod go.mod

RUN go mod download
RUN go build -x -o media-roller ./src

# youtube-dl needs python
FROM python:3.8.1-alpine3.11
RUN apk add --no-cache ffmpeg \
    curl && \
    ffmpeg -version

COPY --from=builder /app/media-roller /app/media-roller
COPY templates /app/templates

WORKDIR /app

RUN curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/local/bin/youtube-dl && \
   chmod a+rx /usr/local/bin/youtube-dl && \
   youtube-dl --version

CMD /app/media-roller
