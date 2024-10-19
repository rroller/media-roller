#!/usr/bin/env bash
docker run -p 3000:3000 -v $(pwd)/download:/download media-roller
